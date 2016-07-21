package controller

import (
	"gosaic/environment"
	"gosaic/model"
	"gosaic/util"
)

type addIndex struct {
	path   string
	md5sum string
}

func Index(env environment.Environment, path string) {
	paths, err := getJpgPaths(path, env)
	if err != nil {
		env.Printf("Error finding images in path %s: %s\n", path, err.Error())
		return
	}

	num := len(paths)
	if num == 0 {
		env.Printf("No images found at path %s\n", path)
		return
	}

	env.Printf("Processing %d images\n", num)
	err = processIndexPaths(paths, env)
	if err != nil {
		env.Printf("Error indexing images: %s\n", err.Error())
	}
}

func processIndexPaths(paths []string, env environment.Environment) error {
	add := make(chan addIndex)
	sem := make(chan bool, env.Workers())

	go storeIndexPaths(add, sem, env)

	for _, p := range paths {
		sem <- true
		go func(myPath string) {
			md5sum, err := util.Md5sum(myPath)
			if err != nil {
				env.Printf("Unable to get md5 sum for path %s\n", myPath)
				return
			}
			add <- addIndex{myPath, md5sum}
		}(p)
	}

	for i := 0; i < cap(sem); i++ {
		sem <- true
	}

	return nil
}

func storeIndexPaths(add <-chan addIndex, sem <-chan bool, env environment.Environment) {
	for newIndex := range add {
		storeIndexPath(newIndex, env)
		<-sem
	}
}

func storeIndexPath(newIndex addIndex, env environment.Environment) {
	gidxService, err := env.GidxService()
	if err != nil {
		env.Println(err.Error())
		return
	}

	aspectService, err := env.AspectService()
	if err != nil {
		env.Println(err.Error())
		return
	}

	exists, err := gidxService.ExistsBy("md5sum", newIndex.md5sum)
	if err != nil {
		env.Println("Failed to lookup md5sum", newIndex.md5sum, err)
		return
	}

	if exists {
		return
	}

	env.Println(newIndex.path)

	img, err := util.OpenImage(newIndex.path)
	if err != nil {
		env.Println("Can't open image", newIndex.path, err)
		return
	}

	// don't actually fix orientation here, just determine
	// if x and y need to be swapped
	orientation, err := util.GetOrientation(newIndex.path)
	swap := false
	if err == nil && 4 < orientation && orientation <= 8 {
		swap = true
	}
	if orientation == 0 {
		orientation = 1
	}

	bounds := (*img).Bounds()

	var width, height int
	if swap {
		width = bounds.Max.Y
		height = bounds.Max.X
	} else {
		width = bounds.Max.X
		height = bounds.Max.Y
	}

	aspect, err := aspectService.FindOrCreate(width, height)
	if err != nil {
		env.Println("Error getting image aspect data", newIndex.path, err)
		return
	}

	gidx := model.NewGidx(aspect.Id, newIndex.path, newIndex.md5sum, uint(width), uint(height), orientation)
	err = gidxService.Insert(gidx)
	if err != nil {
		env.Println("Error storing image data", newIndex.path, err)
	}
}