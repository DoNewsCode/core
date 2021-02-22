/*
Package ots3 provides a s3 uploader with opentracing capabilities.

Introduction

S3 is the de facto standard for cloud file systems.The transport of S3 is HTTP(s) which
is pleasantly simple to trace. This package also features a go kit server and client for the
upload service.

Simple Usage

Creating a s3 manager:

	var manager = NewManager(accessKey, accessSecret, endpoint, region, bucket)
	url, err := manager.Upload(context.Background(), "myfile", file)

Integration

Package ots3 exports the following configuration:

	s3:
	  default:
	    accessKey:
	    accessSecret:
	    bucket:
	    endpoint:
	    region:
	    cdnUrl:

To use package ots3 with package core:

	var c *core.C = core.New()
	c.Provide(mods3.Provide)
	c.AddModuleFunc(mods3.New)
	c.Invoke(function(manager *ots3.Manager) {
		// do something with manager
	})

Adding the module created by mods3.New is optional. This module provides an "/upload" path
for the http router. If this is not relevant, just leave it out.

Sometimes there are valid reasons to connect to more than one s3 server. Inject
mods3.S3Maker to factory a *ots3.Manager with a specific configuration entry.

	c.Invoke(function(maker mods3.S3Maker) {
		manager, err := maker.Make("default")
	})

Future scope

Currently this package only focus on the file upload aspect of s3. Other s3 features can be
incrementally implemented.
*/
package ots3
