/*
 * This file is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License, version 2,
 * as published by the Free Software Foundation.
 *
 * In addition to the permissions in the GNU General Public License,
 * the authors give you unlimited permission to link the compiled
 * version of this file into combinations with other programs,
 * and to distribute those combinations without any restriction
 * coming from the use of this file.  (The General Public License
 * restrictions do apply in other respects; for example, they cover
 * modification of the file, and distribution when not linked into
 * a combined executable.)
 *
 * This file is distributed in the hope that it will be useful, but
 * WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
 * General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program; see the file COPYING.  If not, write to
 * the Free Software Foundation, 51 Franklin Street, Fifth Floor,
 * Boston, MA 02110-1301, USA.
 */

#include <assert.h>
#include <string.h>
#include <git2.h>
#include <git2/odb.h>
#include <git2/refdb.h>
#include <git2/sys/odb_backend.h>
#include <git2/sys/refdb_backend.h>
#include <git2/sys/refs.h>
#include <git2/buffer.h>
#include <libpq-fe.h>
#include <stdlib.h>

#define GIT2_ODB_TABLE_NAME "odb"
#define GIT2_REFDB_TABLE_NAME "refdb"
#define GIT_SYMREF "ref: "

typedef struct {
	git_odb_backend parent;
	PGconn *db;
	char *repo_path;
} postgres_odb_backend;

typedef struct {
	git_refdb_backend parent;
	PGconn *db;
	char *repo_path;
} postgres_refdb_backend;

typedef struct {
	git_reference_iterator parent;

	size_t current;
	PGresult *keys;

	postgres_refdb_backend *backend;
} postgres_refdb_iterator;

/*
 * ODB -> Objects DB functions
 */

// static int init_odb(PGconn *db)
// {
// 	PGresult *result;
	
// 	static const char *sql_creat =
// 		"CREATE TABLE IF NOT EXISTS " GIT2_ODB_TABLE_NAME " ("
// 		"oid TEXT NOT NULL,"
// 		"type INTEGER NOT NULL,"
// 		"size INTEGER NOT NULL,"
// 		"data bytea,"
// 		"repo TEXT NOT NULL);";
	
// 	result = PQexec(db, sql_creat);
// 	if(PQresultStatus(result) != PGRES_COMMAND_OK && PQresultStatus(result) != PGRES_TUPLES_OK){
// 		return GIT_ERROR;
// 	}

// 	return GIT_OK;
// }

int postgres_odb_backend__read_header(size_t *len_p, git_otype *type_p, git_odb_backend *_backend, const git_oid *oid)
{
	postgres_odb_backend *backend;
	int error;
	PGresult *result;

	assert(len_p && type_p && _backend && oid);

	backend = (postgres_odb_backend *) _backend;
	error = GIT_ERROR;

	char *str_id = calloc(GIT_OID_HEXSZ + 1, sizeof(char));
	git_oid_tostr(str_id, GIT_OID_HEXSZ, oid);
    const char *paramValues[2] = {str_id, backend->repo_path};
	
	result = PQexecParams(backend->db, "SELECT type, size FROM " GIT2_ODB_TABLE_NAME " WHERE oid = $1 and repo = $2;", 2, NULL, paramValues, NULL, NULL, 0);
	if(PQresultStatus(result) != PGRES_TUPLES_OK){
		return GIT_ERROR;
	}
	
	if(PQntuples(result) < 1){
		error = GIT_ENOTFOUND;
	}
	else{
		assert(PQntuples(result) == 1);
		
		*type_p = (git_otype)strtol(PQgetvalue(result, 0, 0), NULL, 10);
		*len_p = (git_otype)strtol(PQgetvalue(result, 0, 1), NULL, 10);
		error = GIT_OK;
	}

	PQclear(result);
	return error;
}

int postgres_odb_backend__read(void **data_p, size_t *len_p, git_otype *type_p, git_odb_backend *_backend, const git_oid *oid)
{
	postgres_odb_backend *backend;
	int error;
	PGresult *result;

	assert(data_p && len_p && type_p && _backend && oid);

	backend = (postgres_odb_backend *) _backend;
	error = GIT_ERROR;

	char *str_id = calloc(GIT_OID_HEXSZ + 1, sizeof(char));
	git_oid_tostr(str_id, GIT_OID_HEXSZ, oid);
    const char *paramValues[2] = {str_id, backend->repo_path};

	result = PQexecParams(backend->db, "SELECT type, size, data FROM " GIT2_ODB_TABLE_NAME " WHERE oid = $1 and repo = $2;", 2, NULL, paramValues, NULL, NULL, 0);
	if(PQresultStatus(result) != PGRES_TUPLES_OK){
		return GIT_ERROR;
	}
	
	if(PQntuples(result) < 1){
		error = GIT_ENOTFOUND;
	}
	else{
		assert(PQntuples(result) == 1);
		
		*type_p = (git_otype)strtol(PQgetvalue(result, 0, 0), NULL, 10);
		// *len_p = (size_t)strtol(PQgetvalue(result, 0, 1), NULL, 10);
		// *data_p = malloc(*len_p);
		*data_p =  PQunescapeBytea((unsigned char *)PQgetvalue(result, 0, 2), len_p);
		if (data_p == NULL) {
			error = GIT_ERROR_NOMEMORY;
		} else {
			// memcpy(*data_p, data, *len_p);
			error = GIT_OK;
		}
		
		error = GIT_OK;
	}

	PQclear(result);
	return error;
}

int postgres_odb_backend__read_prefix(git_oid *out_oid, void **data_p, size_t *len_p, git_otype *type_p, git_odb_backend *_backend,
					const git_oid *short_oid, unsigned int len) {
	if (len >= GIT_OID_HEXSZ) {
		/* Just match the full identifier */
		int error = postgres_odb_backend__read(data_p, len_p, type_p, _backend, short_oid);
		if (error == GIT_OK)
			git_oid_cpy(out_oid, short_oid);

		return error;
	} else if (len < GIT_OID_HEXSZ) {
		return GIT_ERROR;
	}
	return GIT_OK;
}

int postgres_odb_backend__exists(git_odb_backend *_backend, const git_oid *oid)
{
	postgres_odb_backend *backend;
	int found;
	PGresult *result;

	assert(_backend && oid);

	backend = (postgres_odb_backend *) _backend;
	found = 0;

	char *str_id = calloc(GIT_OID_HEXSZ + 1, sizeof(char));
	git_oid_tostr(str_id, GIT_OID_HEXSZ, oid);
    const char *paramValues[2] = {str_id, backend->repo_path};
	
	result = PQexecParams(backend->db, "SELECT type, size, data FROM " GIT2_ODB_TABLE_NAME " WHERE oid = $1 and repo = $2;", 2, NULL, paramValues, NULL, NULL, 0);
	if(PQresultStatus(result) != PGRES_TUPLES_OK){
		return GIT_ERROR;
	}

	if(PQntuples(result) > 0){
		found = 1;
	}

	PQclear(result);
	return found;
}

int postgres_odb_backend__write(git_odb_backend *_backend, const git_oid *oid, const void *data, size_t len, git_otype type)
{
	int error;
	postgres_odb_backend *backend;
	PGresult *result;
	
	assert(oid && _backend && data);

	backend = (postgres_odb_backend *) _backend;
	
	//this is a rather ugly construct to avoid having to know about postgres' internal integer representation
	char type_str[128];
	char size_str[128];

	char *str_id = calloc(GIT_OID_HEXSZ + 1, sizeof(char));
	git_oid_tostr(str_id, GIT_OID_HEXSZ, oid);
	
	const char *values[5] = {str_id, type_str, size_str, data, backend->repo_path};
	const int lengths[5] = {0, 0, 0, len, 0};
	const int formats[5] = {0, 0, 0, 1, 0};

	// if ((error = git_odb_hash(id, data, len, type)) < 0)
	// 	return error;

	snprintf(type_str, sizeof(type_str), "%d", type);
	snprintf(size_str, sizeof(size_str), "%lu", len);
	
	result = PQexecParams(backend->db, "INSERT INTO " GIT2_ODB_TABLE_NAME " VALUES ($1, $2, $3, $4, $5);", 5, NULL, values, lengths, formats, 0);

	error = PQresultStatus(result);
	PQclear(result);
	
	return (error == PGRES_COMMAND_OK || error == PGRES_TUPLES_OK) ? GIT_OK : GIT_ERROR;
}

void postgres_odb_backend__free(git_odb_backend *_backend)
{
	postgres_odb_backend *backend;
	assert(_backend);
	backend = (postgres_odb_backend *) _backend;

	free(backend->repo_path);

	free(backend);
}

/*
 * REFDB -> Reference DB functions
 */

// static int init_refdb(PGconn *db)
// {
// 	PGresult *result;
	
// 	static const char *sql_creat =
// 		"CREATE TABLE IF NOT EXISTS " GIT2_REFDB_TABLE_NAME " ("
// 		"repo    TEXT NOT NULL,"
// 		"refname TEXT NOT NULL,"
// 		"target  TEXT NOT NULL,"
// 		"type INTEGER NOT NULL,"
// 		"PRIMARY KEY (repo, refname));";
	
// 	result = PQexec(db, sql_creat);
// 	if(PQresultStatus(result) != PGRES_COMMAND_OK && PQresultStatus(result) != PGRES_TUPLES_OK){
// 		return GIT_ERROR;
// 	}

// 	return GIT_OK;
// }

int postgres_refdb_backend__exists(int *exists, git_refdb_backend *_backend, const char *ref_name)
{
  postgres_refdb_backend *backend = (postgres_refdb_backend *)_backend;
  PGresult *result;

  assert(backend);

  *exists = 0;
	
	const char* paramValues[2] = {backend->repo_path, ref_name};
  result = PQexecParams(backend->db, "SELECT target FROM " GIT2_REFDB_TABLE_NAME " WHERE refname = $1 and repo = $2;", 2, NULL, paramValues, NULL, NULL, 0);

  if (PQresultStatus(result) != PGRES_TUPLES_OK) {
    return GIT_ERROR;
  }
	
	if (PQntuples(result) > 0) {
		*exists = 1;
	}

  PQclear(result);
  return GIT_OK;
}

int postgres_refdb_backend__lookup(git_reference **out, git_refdb_backend *_backend, const char *ref_name)
{
  postgres_refdb_backend *backend = (postgres_refdb_backend *)_backend;
	int error = GIT_OK;
	PGresult *result;
	git_oid oid;

	assert(ref_name && _backend);

	backend = (postgres_refdb_backend *) _backend;

	const char* paramValues[2] = {ref_name, backend->repo_path};
  	result = PQexecParams(backend->db, "SELECT type, target FROM " GIT2_REFDB_TABLE_NAME " WHERE refname = $1 and repo = $2;", 2, NULL, paramValues, NULL, NULL, 0);

	if (PQresultStatus(result) != PGRES_TUPLES_OK) {
		error = GIT_ERROR;
	} else {
		if (PQntuples(result) > 0) {
			git_ref_t type = (git_ref_t) atoi(PQgetvalue(result, 0, 0));

			if (out == NULL && (type == GIT_REF_OID || type == GIT_REF_SYMBOLIC)){
				return GIT_OK;
			}

			if (type == GIT_REF_OID) {
				git_oid_fromstr(&oid, PQgetvalue(result, 0, 1));
				*out = git_reference__alloc(ref_name, &oid, NULL);
			} else if (type == GIT_REF_SYMBOLIC) {
				*out = git_reference__alloc_symbolic(ref_name, PQgetvalue(result, 0, 1));
			} else {
				error = GIT_ERROR;
			}
		} else {
			error = GIT_ENOTFOUND;
		}
	}

	PQclear(result);
	return error;
}

int postgres_refdb_backend__iterator_next(git_reference **ref, git_reference_iterator *_iter) {
	postgres_refdb_iterator *iter;
	postgres_refdb_backend *backend;
	char* ref_name;
	int error;

	assert(_iter);
	iter = (postgres_refdb_iterator *) _iter;

	if(iter->current < PQntuples(iter->keys)) {
		ref_name = PQgetvalue(iter->keys, iter->current++, 0);
		// ref_name = strstr(iter->keys->element[iter->current++]->str, ":refdb:") + 7;
		error = postgres_refdb_backend__lookup(ref, (git_refdb_backend *) iter->backend, ref_name);

		return error;
	} else {
		return GIT_ITEROVER;
	}
}

int postgres_refdb_backend__iterator_next_name(const char **ref_name, git_reference_iterator *_iter) {
	postgres_refdb_iterator *iter;

	assert(_iter);
	iter = (postgres_refdb_iterator *) _iter;

	if(iter->current < PQntuples(iter->keys)) {
		char *ref_name = strdup(PQgetvalue(iter->keys, iter->current++, 0));
		// *ref_name = strdup(strstr(iter->keys->element[iter->current++]->str, ":refdb:") + 7);
		return GIT_OK;
	} else {
		return GIT_ITEROVER;
	}
}

void postgres_refdb_backend__iterator_free(git_reference_iterator *_iter) {
	postgres_refdb_iterator *iter;

	assert(_iter);
	iter = (postgres_refdb_iterator *) _iter;

	PQclear(iter->keys);

	free(iter);
}

int postgres_refdb_backend__iterator(git_reference_iterator **_iter, git_refdb_backend *_backend, const char *glob)
{
	postgres_refdb_backend *backend;
	postgres_refdb_iterator *iterator;
	int error = GIT_OK;
	PGresult *result;

	assert(_backend);

	backend = (postgres_refdb_backend *) _backend;

	const char* paramValues[2] = {(glob != NULL ? glob : "refs/%%"), backend->repo_path};
  result = PQexecParams(backend->db, "SELECT refname FROM " GIT2_REFDB_TABLE_NAME " WHERE refname LIKE $1 and repo = $2;", 2, NULL, paramValues, NULL, NULL, 0);
	if (PQresultStatus(result) != PGRES_TUPLES_OK) {
		PQclear(result);
		giterr_set_str(GITERR_REFERENCE, "Postgres refdb storage error");
		return GIT_ERROR;
	}

	iterator = calloc(1, sizeof(postgres_refdb_iterator));

	iterator->backend = backend;
	iterator->keys = result;

	iterator->parent.next = &postgres_refdb_backend__iterator_next;
	iterator->parent.next_name = &postgres_refdb_backend__iterator_next_name;
	iterator->parent.free = &postgres_refdb_backend__iterator_free;

	*_iter = (git_reference_iterator *) iterator;

	return GIT_OK;
}

int postgres_refdb_backend__write(git_refdb_backend *_backend, const git_reference *ref, int force, const git_signature *who, const char *message, const git_oid *old, const char *old_target)
{
	postgres_refdb_backend *backend;
	int error = GIT_OK;
	PGresult *result;

	const char *name = git_reference_name(ref);
	const git_oid *target;
	const char *symbolic_target;
	char oid_str[GIT_OID_HEXSZ + 1];
	char type_str[128];

	assert(ref && _backend);

	backend = (postgres_refdb_backend *) _backend;
	target = git_reference_target(ref);
	symbolic_target = git_reference_symbolic_target(ref);

	/* FIXME handle force correctly */

	const char *values[4] = {backend->repo_path, name, symbolic_target, type_str};
	const int lengths[4] = {0, 0, 0, 0};
	const int formats[4] = {0, 0, 0, 0};

	if (target) {
		git_oid_nfmt(oid_str, sizeof(oid_str), target);
		values[2] = oid_str;
		values[3] = type_str;
		snprintf(type_str, sizeof(type_str), "%d", GIT_REF_OID);
	} else {
		values[2] = symbolic_target;
		values[3] = type_str;
		snprintf(type_str, sizeof(type_str), "%d", GIT_REF_SYMBOLIC);
	}
	
	// Kinda sucks that we have to do a lookup for each write. Would be better to detect the unique key constraint error and attempt an update instead.
	if (postgres_refdb_backend__lookup(NULL, _backend, name) < 0) {
			result = PQexecParams(backend->db, "INSERT INTO " GIT2_REFDB_TABLE_NAME " VALUES ($1, $2, $3, $4);", 4, NULL, values, lengths, formats, 0);
	} else {
			result = PQexecParams(backend->db, "UPDATE " GIT2_REFDB_TABLE_NAME " SET repo = $1, refname = $2, target = $3, type = $4 WHERE repo = $1 and refname = $2;", 4, NULL, values, lengths, formats, 0);
	}

  	if (PQresultStatus(result) != PGRES_COMMAND_OK) {
    	error = GIT_ERROR;
  	} else {
    	error = GIT_OK;
  	}

	PQclear(result);
	return error;
}

int postgres_refdb_backend__delete(git_refdb_backend *_backend, const char *ref_name, const git_oid *old, const char *old_target)
{
	postgres_refdb_backend *backend;
	int error = GIT_OK;
	PGresult *result;

	assert(ref_name && _backend);

	backend = (postgres_refdb_backend *) _backend;

	const char *values[2] = {backend->repo_path, ref_name};
	const int lengths[2] = {0, 0};
	const int formats[2] = {0, 0};

	result = PQexecParams(backend->db, "DELETE FROM " GIT2_REFDB_TABLE_NAME " WHERE repo = $1 and refname = $2;", 2, NULL, values, lengths, formats, 0);

	if (PQresultStatus(result) != PGRES_COMMAND_OK) {
    	error = GIT_ERROR;
  	} else {
    	error = GIT_OK;
  	}

	PQclear(result);
	return error;
}

int postgres_refdb_backend__rename(git_reference **out, git_refdb_backend *_backend, const char *old_name, const char *new_name, int force, const git_signature *who, const char *message)
{
	postgres_refdb_backend *backend;
	int error = GIT_OK;
	PGresult *result;

	assert(old_name && new_name && _backend);

	backend = (postgres_refdb_backend *) _backend;

	const char *values[3] = {backend->repo_path, old_name, new_name};
	const int lengths[3] = {0, 0, 0};
	const int formats[3] = {0, 0, 0};

	result = PQexecParams(backend->db, "UPDATE " GIT2_REFDB_TABLE_NAME " SET refname = $3 WHERE repo = $1 and refname = $2;", 3, NULL, values, lengths, formats, 0);

	if (PQresultStatus(result) != PGRES_COMMAND_OK) {
    	error = GIT_ERROR;
  	} else {
    	error = GIT_OK;
  	}

	PQclear(result);
	return error;
}

void postgres_refdb_backend__free(git_refdb_backend *_backend)
{
	postgres_refdb_backend *backend;

	assert(_backend);
	backend = (postgres_refdb_backend *) _backend;

	free(backend->repo_path);

	free(backend);
}

/* reflog methods */

int postgres_refdb_backend__has_log(git_refdb_backend *_backend, const char *refname)
{
	return 0;
}

int postgres_refdb_backend__ensure_log(git_refdb_backend *_backend, const char *refname)
{
	return GIT_ERROR;
}

int postgres_refdb_backend__reflog_read(git_reflog **out, git_refdb_backend *_backend, const char *name)
{
	return GIT_ERROR;
}

int postgres_refdb_backend__reflog_write(git_refdb_backend *_backend, git_reflog *reflog)
{
	return GIT_ERROR;
}

int postgres_refdb_backend__reflog_rename(git_refdb_backend *_backend, const char *old_name, const char *new_name)
{
	return GIT_ERROR;
}

int postgres_refdb_backend__reflog_delete(git_refdb_backend *_backend, const char *name)
{
	return GIT_ERROR;
}


/* External Constructor Functions */

int git_odb_backend_postgres(git_odb_backend **backend_out, const char* path, const char *pg_host,
        const char *pg_user, const char *pg_passwd, const char *pg_db, unsigned int pg_port)
{
	postgres_odb_backend *backend;
	int error;
	PGconn *sharedDbConn;

	backend = calloc(1, sizeof(postgres_odb_backend));
	if (backend == NULL)
		return GITERR_NOMEMORY;

	char port_str[10];
	snprintf(port_str, sizeof(port_str), "%d", pg_port);	

	//This allows the application to use the .pgpass mechanism by supplying a NULL password
	char const *keywords[] = {"host", "port", "dbname", "user",  (pg_passwd) ? "password":NULL,  NULL};
	const char *values[] = {pg_host, port_str, pg_db, pg_user, pg_passwd, NULL};
	sharedDbConn = PQconnectdbParams(keywords, (char const**)values, 0);
	if(PQstatus(sharedDbConn) != CONNECTION_OK){
		PQfinish(sharedDbConn);
		goto cleanup;
	}

	// check for and possibly create the database
	// error = init_odb(sharedDbConn);
	// if (error < 0)
	//	 goto cleanup;

	backend->db = sharedDbConn;
	backend->repo_path = strdup(path);
	backend->parent.version = GIT_ODB_BACKEND_VERSION;

	backend->parent.read = &postgres_odb_backend__read;
	backend->parent.read_header = &postgres_odb_backend__read_header;
	backend->parent.write = &postgres_odb_backend__write;
	backend->parent.exists = &postgres_odb_backend__exists;
	backend->parent.free = &postgres_odb_backend__free;

	backend->parent.writestream = NULL;
	backend->parent.foreach = NULL;

	*backend_out = (git_odb_backend *) backend;
	return GIT_OK;

cleanup:
	postgres_odb_backend__free((git_odb_backend *)backend);
	return GIT_ERROR;
}

int git_refdb_backend_postgres(git_refdb_backend **backend_out, const char* path, const char *pg_host,
        const char *pg_user, const char *pg_passwd, const char *pg_db, unsigned int pg_port)
{
	postgres_refdb_backend *backend;
	int error;
	PGconn *sharedDbConn;

	backend = calloc(1, sizeof(postgres_refdb_backend));
	if (backend == NULL)
		return GITERR_NOMEMORY;

	char port_str[10];
	snprintf(port_str, sizeof(port_str), "%d", pg_port);	

	//This allows the application to use the .pgpass mechanism by supplying a NULL password
	char const *keywords[] = {"host", "port", "dbname", "user",  (pg_passwd) ? "password":NULL,  NULL};
	const char *values[] = {pg_host, port_str, pg_db, pg_user, pg_passwd, NULL};
	sharedDbConn = PQconnectdbParams(keywords, (char const**)values, 0);
	if(PQstatus(sharedDbConn) != CONNECTION_OK){
		PQfinish(sharedDbConn);
		goto cleanup;
	}

	// check for and possibly create the database
	// error = init_refdb(sharedDbConn);
	// if (error < 0)
	//	 goto cleanup;

	backend->db = sharedDbConn;
	backend->repo_path = strdup(path);
	backend->parent.version = GIT_REFDB_BACKEND_VERSION;

	backend->parent.exists = &postgres_refdb_backend__exists;
	backend->parent.lookup = &postgres_refdb_backend__lookup;
	backend->parent.iterator = &postgres_refdb_backend__iterator;
	backend->parent.write = &postgres_refdb_backend__write;
	backend->parent.del = &postgres_refdb_backend__delete;
	backend->parent.rename = &postgres_refdb_backend__rename;
	backend->parent.compress = NULL;
	backend->parent.free = &postgres_refdb_backend__free;

	backend->parent.has_log = &postgres_refdb_backend__has_log;
	backend->parent.ensure_log = &postgres_refdb_backend__ensure_log;
	backend->parent.reflog_read = &postgres_refdb_backend__reflog_read;
	backend->parent.reflog_write = &postgres_refdb_backend__reflog_write;
	backend->parent.reflog_rename = &postgres_refdb_backend__reflog_rename;
	backend->parent.reflog_delete = &postgres_refdb_backend__reflog_delete;

	*backend_out = (git_refdb_backend *) backend;

	return GIT_OK;

cleanup:
	postgres_refdb_backend__free((git_refdb_backend *)backend);
	return GIT_ERROR;
}

void git_refdb_backend_free_postgres(git_refdb_backend *_backend)
{
	postgres_refdb_backend *backend;

	assert(_backend);
	backend = (postgres_refdb_backend *) _backend;

	PQfinish(backend->db);
}

void git_odb_backend_free_postgres(git_odb_backend *_backend)
{
	postgres_odb_backend *backend;
	assert(_backend);
	backend = (postgres_odb_backend *) _backend;

	PQfinish(backend->db);
}