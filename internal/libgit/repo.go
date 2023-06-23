package libgit

/*
#cgo pkg-config: libgit2
#cgo CFLAGS: -DLIBGIT2_DYNAMIC
#cgo LDFLAGS: -L ../../libgit2-backends/build -lgit2-postgres
#include <git2.h>
#include <git2/sys/odb_backend.h>

extern int git_odb_backend_postgres(git_odb_backend **backend_out, const char* path, const char *pg_host,
	const char *pg_user, const char *pg_passwd, const char *pg_db, unsigned int pg_port);
extern int git_refdb_backend_postgres(git_refdb_backend **backend_out, const char* path, const char *pg_host,
	const char *pg_user, const char *pg_passwd, const char *pg_db, unsigned int pg_port);
extern void git_odb_backend_free_postgres(git_odb_backend *_backend);
extern void git_refdb_backend_free_postgres(git_refdb_backend *_backend);
*/
import "C"

import (
	"bipgit/internal/configs"
	"bipgit/internal/constants"
	"errors"
	"fmt"
	"time"
	"unsafe"

	git "github.com/libgit2/git2go/v33"
)

type BipRepo struct {
	Repo         *git.Repository
	Odb          *git.Odb
	Refdb        *git.Refdb
	OdbBackend   *C.git_odb_backend
	RefdbBackend *C.git_refdb_backend
}

func InitRepo(repoPath string, userName string, userEmail string) (*BipRepo, *git.Signature) {
	repo := initWithPostgresBackend(repoPath)
	signature := &git.Signature{
		Name:  userName,
		Email: userEmail,
		When:  time.Now(),
	}
	return repo, signature
}

func initPostgresOdbBackend(repoPath string) (*git.Odb, *C.git_odb_backend) {
	odb, err := git.NewOdb()
	if err != nil {
		fmt.Println(errors.New("1st Panic - initPostgresOdbBackend: " + err.Error()))
		panic(err)
	}
	var odbBackendC *C.git_odb_backend = nil
	errCode := C.git_odb_backend_postgres(&odbBackendC,
		C.CString(repoPath),
		C.CString(configs.GetDBHost()),
		C.CString(configs.GetDBUser()),
		C.CString(configs.GetDBPassword()),
		C.CString(configs.GetDBName()),
		C.uint(configs.GetDBPort()),
	)
	if errCode != 0 {
		panic(errors.New("Failed connecting to Postgres"))
	}
	backend := git.NewOdbBackendFromC(unsafe.Pointer(odbBackendC))
	err = odb.AddBackend(backend, 1)
	if err != nil {
		fmt.Println(errors.New("2nd Panic - initPostgresOdbBackend: " + err.Error()))
		panic(err)
	}
	return odb, odbBackendC
}

func initPostgresRefdbBackend(repo *git.Repository, repoPath string) (*git.Refdb, *C.git_refdb_backend) {
	refdb, err := repo.NewRefdb()
	if err != nil {
		fmt.Println(errors.New("1st Panic - initPostgresRefdbBackend: " + err.Error()))
		panic(err)
	}
	var refdbBackendC *C.git_refdb_backend = nil
	errCode := C.git_refdb_backend_postgres(&refdbBackendC,
		C.CString(repoPath),
		C.CString(configs.GetDBHost()),
		C.CString(configs.GetDBUser()),
		C.CString(configs.GetDBPassword()),
		C.CString(configs.GetDBName()),
		C.uint(configs.GetDBPort()),
	)
	if errCode != 0 {
		panic(errors.New("Failed connecting to Postgres"))
	}
	backend := git.NewRefdbBackendFromC(unsafe.Pointer(refdbBackendC))
	err = refdb.SetBackend(backend)
	if err != nil {
		fmt.Println(errors.New("2nd Panic - initPostgresRefdbBackend: " + err.Error()))
		panic(err)
	}
	repo.SetRefdb(refdb)
	return refdb, refdbBackendC
}

func initWithPostgresBackend(repoPath string) *BipRepo {
	odb, obackend := initPostgresOdbBackend(repoPath)
	repo, err := git.NewRepositoryWrapOdb(odb)
	if err != nil {
		fmt.Println(errors.New("1st Panic - initWithPostgresBackend: " + err.Error()))
		panic(err)
	}
	refdb, rbackend := initPostgresRefdbBackend(repo, repoPath)

	var head *git.Reference
	head, err = repo.References.Lookup("HEAD")
	if err != nil {
		if gitErr, isOK := err.(*git.GitError); isOK && gitErr.Code == git.ErrNotFound {
			head, err = repo.References.CreateSymbolic("HEAD", "refs/heads/"+constants.DefaultBranchName, false, "")
			if err != nil {
				fmt.Println(errors.New("2nd Panic - initWithPostgresBackend: " + err.Error()))
				panic(err)
			}
		} else {
			fmt.Println(errors.New("3rd Panic - initWithPostgresBackend: " + err.Error()))
			panic(err)
		}
	}
	if err == nil {
		head.Free()
	}

	return &BipRepo{
		Repo:         repo,
		Odb:          odb,
		Refdb:        refdb,
		OdbBackend:   obackend,
		RefdbBackend: rbackend,
	}
}

func (bipRepo *BipRepo) CloseConn() {
	C.git_odb_backend_free_postgres(bipRepo.OdbBackend)
	C.git_refdb_backend_free_postgres(bipRepo.RefdbBackend)
}
