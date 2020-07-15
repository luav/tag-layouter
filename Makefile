tag-layouter: apriltag/libapriltag.a oldtags/liboldtags.a
	go test -coverprofile cover.out
	go build

apriltag/libapriltag.a:
	# -I to add additional include path for the latest OpenCV versions
	$(MAKE) -C apriltag -I /usr/include/opencv4

oldtags/liboldtags.a:
	$(MAKE) -C oldtags
