docs: 
	-rm docs/*.svg
	./go-cover-treemap -coverprofile testdata/treemap.cover > docs/go-cover-treemap.svg
	./go-cover-treemap -coverprofile testdata/go-featureprocessing.cover > docs/go-featureprocessing.svg
	./go-cover-treemap -coverprofile testdata/gin.cover > docs/gin.svg
	./go-cover-treemap -coverprofile testdata/chi.cover > docs/chi.svg
	./go-cover-treemap -coverprofile testdata/hugo.cover > docs/hugo.svg

.PHONY: docs
