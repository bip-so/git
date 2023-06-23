PROJECT(libgit2-postgres C)
CMAKE_MINIMUM_REQUIRED(VERSION 2.6)

INCLUDE(../CMake/FindLibgit2.cmake)
INCLUDE(../CMake/FindLibpq.cmake)

# Build options
OPTION (BUILD_SHARED_LIBS "Build Shared Library (OFF for Static)" ON)
OPTION (BUILD_TESTS "Build Tests" ON)

# Build Release by default
IF (NOT CMAKE_BUILD_TYPE)
    SET(CMAKE_BUILD_TYPE "Release" CACHE STRING "Choose the type of build, options are: Debug Release RelWithDebInfo MinSizeRel." FORCE)
ENDIF ()

# Compile and link libgit2
INCLUDE_DIRECTORIES(${LIBGIT2_INCLUDE_DIR} ${LIBPQ_INCLUDE_DIR})

IF (BUILD_SHARED_LIBS)
    ADD_LIBRARY(git2-postgres SHARED postgres.c)
ELSE ()
    ADD_LIBRARY(git2-postgres STATIC postgres.c)
ENDIF ()

TARGET_LINK_LIBRARIES(git2-postgres ${LIBGIT2_LIBRARIES} ${LIBPQ_LIBRARIES})
