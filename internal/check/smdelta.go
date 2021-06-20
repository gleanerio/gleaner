package check

import "fmt"

// NewURLs is called and looks for a parquet file for this organization.
// It will then do an s3select to pull the URLs from the current loaded graph
// and compare these to the URLs from the sitemap.  The diff represents the
// resources to index and are passed to summoner.
func NewURLs() []string {

	fmt.Println("check URLs from sitemap vs those in prov file org-[latest].parquet")

	return []string{"url list", "to reutrn"}

}
