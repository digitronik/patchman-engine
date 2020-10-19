package controllers

import (
	"app/base/database"
	"app/manager/middlewares"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

var PackagesFields = database.MustGetQueryAttrs(&PackageItem{})
var PackagesSelect = database.MustGetSelect(&PackageItem{})
var PackagesOpts = ListOpts{
	Fields: PackagesFields,
	// By default, we show only fresh systems. If all systems are required, you must pass in:true,false filter into the api
	DefaultFilters: map[string]FilterData{},
	DefaultSort:    "name",
}

type PackageItem struct {
	Name             string `json:"name" query:"res.name"`
	SystemsInstalled int    `json:"systems_installed" query:"res.systems_installed"`
	SystemsUpdatable int    `json:"systems_updatable" query:"res.systems_updatable"`
	Summary          string `json:"summary" query:"latest.summary"`
}

type PackagesResponse struct {
	Data  []PackageItem `json:"data"`
	Links Links         `json:"links"`
	Meta  ListMeta      `json:"meta"`
}

// Used as a for subquery performing the actual calculation which is joined with latest summaries
type queryItem struct {
	NameID           int    `query:"p.name_id"`
	Name             string `json:"name" query:"pn.name"`
	SystemsInstalled int    `json:"systems_installed" query:"count(spkg.system_id)"`
	SystemsUpdatable int    `json:"systems_updatable" query:"count(spkg.system_id) filter (where spkg.updatable)"`
}

var queryItemSelect = database.MustGetSelect(&queryItem{})

// nolint: lll
func packagesQuery(c *gin.Context, acc int) *gorm.DB {
	subQ := database.Db.
		Select(queryItemSelect).
		Table("system_package spkg").
		Joins("inner join system_platform sp on sp.id = spkg.system_id and sp.rh_account_id = ?", acc).
		Where("sp.stale = false").
		Joins("inner join package p on p.id = spkg.package_id").
		Joins("inner join package_name pn on pn.id = p.name_id").
		Where("spkg.rh_account_id = ?", acc).
		Group("p.name_id, pn.name")

	// We need to apply tag filtering on subquery
	subQ, _ = ApplyTagsFilter(c, subQ, "sp.inventory_id")

	return database.Db.
		Select(PackagesSelect).
		Table("package_latest_cache latest").
		Joins("INNER JOIN ? res ON res.name_id = latest.name_id", subQ.SubQuery())
}

// @Summary Show me all installed packages across my systems
// @Description Show me all installed packages across my systems
// @ID listPackages
// @Security RhIdentity
// @Accept   json
// @Produce  json
// @Param    limit          query      int     false   "Limit for paging, set -1 to return all"
// @Param    offset         query      int     false   "Offset for paging"
// @Param    sort           query      string  false   "Sort field" Enums(id,name,systems_installed,systems_updatable)
// @Param    search         query      string  false   "Find matching text"
// @Param    filter[name]    query     string  false "Filter"
// @Param    filter[systems_installed] query   string  false "Filter"
// @Param    filter[systems_updatable] query   string  false "Filter"
// @Param    filter[summary]           query   string  false "Filter"
// @Param    tags                      query   []string  false "Tag filter"
// @Param    filter[system_profile][sap_system]   query string   false "Filter only SAP systems"
// @Param    filter[system_profile][sap_sids][in] query []string false "Filter systems by their SAP SIDs"
// @Success 200 {object} PackagesResponse
// @Router /api/patch/v1/packages/ [get]
func PackagesListHandler(c *gin.Context) {
	account := c.GetInt(middlewares.KeyAccount)

	query := packagesQuery(c, account)
	query = ApplySearch(c, query, "res.name", "latest.summary")
	query, meta, links, err := ListCommon(query, c, "/packages", PackagesOpts)
	if err != nil {
		// Error handling and setting of result code & content is done in ListCommon
		return
	}

	var systems []PackageItem
	err = query.Scan(&systems).Error
	if err != nil {
		LogAndRespError(c, err, "database error")
		return
	}

	c.JSON(200, PackagesResponse{
		Data:  systems,
		Links: *links,
		Meta:  *meta,
	})
}