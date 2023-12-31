# - Try to find the Lib library
# Once done this will define
#
#  LIBPQ_FOUND - System has PQ
#  LIBPQ_INCLUDE_DIR - The Libpq include directory
#  LIBPQ_LIBRARIES - The libraries needed to use PQ
#  LIBPQ_DEFINITIONS - Compiler switches required for using PQ


# use pkg-config to get the directories and then use these values
# in the FIND_PATH() and FIND_LIBRARY() calls
#FIND_PACKAGE(PkgConfig)
#PKG_SEARCH_MODULE(PC_LIBPQ libpq)

#SET(LIBPQ_DEFINITIONS ${PC_LIBPQ_CFLAGS_OTHER})

#FIND_PATH(LIBPQ_INCLUDE_DIR postgres/postgres_fe.h)

#FIND_LIBRARY(LIBPQ_LIBRARIES NAMES libpq
#   HINTS
#   ${PC_LIBPQ_LIBDIR}
#   ${PC_LIBPQ_LIBRARY_DIRS}
#)

#INCLUDE(FindPackageHandleStandardArgs)
#FIND_PACKAGE_HANDLE_STANDARD_ARGS(LibPq DEFAULT_MSG LIBPQ_LIBRARIES LIBPQ_INCLUDE_DIR)

#MARK_AS_ADVANCED(LIBPQ_INCLUDE_DIR LIBPQ_LIBRARIES)

IF (LIBPQ_INCLUDE_DIR AND LIBPQ_LIBRARIES)
   SET(LIBPQ_FIND_QUIETLY TRUE)
ENDIF (LIBPQ_INCLUDE_DIR AND LIBPQ_LIBRARIES)

# Include dir
FIND_PATH(LIBPQ_INCLUDE_DIR 
	      NAMES libpq-fe.h
	      PATH_SUFFIXES pgsql postgresql
)

FIND_LIBRARY(LIBPQ_LIBRARY 
   NAMES pq
)

INCLUDE(FindPackageHandleStandardArgs)
FIND_PACKAGE_HANDLE_STANDARD_ARGS(LIBPQ DEFAULT_MSG LIBPQ_LIBRARY LIBPQ_INCLUDE_DIR)

IF(LIBPQ_FOUND)
	SET( LIBPQ_LIBRARIES ${LIBPQ_LIBRARY} )
ELSE(LIBPQ_FOUND)
	SET( LIBPQ_LIBRARIES )
ENDIF(LIBPQ_FOUND)

MARK_AS_ADVANCED( LIBPQ_LIBRARY LIBPQ_INCLUDE_DIR )