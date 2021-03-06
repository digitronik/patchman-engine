package controllers

import (
	"app/base/core"
	"app/base/utils"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSystemPackages(t *testing.T) {
	utils.SkipWithoutDB(t)
	core.SetupTestEnvironment()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/00000000-0000-0000-0000-000000000013/packages", nil)
	core.InitRouterWithParams(SystemPackagesHandler, 3, "GET", "/:inventory_id/packages").
		ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	var output SystemPackageResponse
	ParseReponseBody(t, w.Body.Bytes(), &output)
	assert.Len(t, output.Data, 4)
	assert.Equal(t, output.Data[0].Name, "bash")
	assert.Equal(t, len(output.Data[0].Updates), 0)
	assert.Equal(t, output.Data[1].Name, "curl")
	assert.Equal(t, len(output.Data[1].Updates), 0)
	assert.Equal(t, output.Data[2].Name, "firefox")
	assert.Equal(t, len(output.Data[2].Updates), 2)
	assert.Equal(t, output.Data[3].Name, "kernel")
	assert.Equal(t, len(output.Data[3].Updates), 0)
}

func TestPackagesSearch(t *testing.T) {
	utils.SkipWithoutDB(t)
	core.SetupTestEnvironment()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET",
		"/00000000-0000-0000-0000-000000000012/packages?search=kernel", nil)
	core.InitRouterWithParams(SystemPackagesHandler, 3, "GET", "/:inventory_id/packages").
		ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	var output SystemPackageResponse
	ParseReponseBody(t, w.Body.Bytes(), &output)
	assert.Len(t, output.Data, 1)
	assert.Equal(t, output.Data[0].Name, "kernel")
}

func TestNoPackages(t *testing.T) {
	utils.SkipWithoutDB(t)
	core.SetupTestEnvironment()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/00000000-0000-0000-0000-000000000001/packages", nil)
	core.InitRouterWithParams(SystemPackagesHandler, 1, "GET", "/:inventory_id/packages").
		ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestSystemPackagesUpdatableOnly(t *testing.T) {
	utils.SkipWithoutDB(t)
	core.SetupTestEnvironment()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET",
		"/00000000-0000-0000-0000-000000000013/packages?filter[updatable]=true", nil)
	core.InitRouterWithParams(SystemPackagesHandler, 3, "GET", "/:inventory_id/packages").
		ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	var output SystemPackageResponse
	ParseReponseBody(t, w.Body.Bytes(), &output)
	assert.Len(t, output.Data, 1)
	assert.Equal(t, output.Data[0].Name, "firefox")
}

func TestSystemPackagesNonUpdatableOnly(t *testing.T) {
	utils.SkipWithoutDB(t)
	core.SetupTestEnvironment()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET",
		"/00000000-0000-0000-0000-000000000013/packages?filter[updatable]=false", nil)
	core.InitRouterWithParams(SystemPackagesHandler, 3, "GET", "/:inventory_id/packages").
		ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	var output SystemPackageResponse
	ParseReponseBody(t, w.Body.Bytes(), &output)
	assert.Len(t, output.Data, 3)
	assert.Equal(t, output.Data[0].Name, "bash")
	assert.Equal(t, output.Data[1].Name, "curl")
	assert.Equal(t, output.Data[2].Name, "kernel")
}

func TestSystemPackagesWrongOffset(t *testing.T) {
	doTestWrongOffset(t, "/:inventory_id/packages",
		"/00000000-0000-0000-0000-000000000001/packages?offset=1000", SystemPackagesHandler)
}
