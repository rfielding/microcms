GoSQLite
=======

This is an experiment in quickly creatding a CMS, which I mean

- upload and download files
- large media files, such as multi-GB mp4s work well and efficiently
- install tarballs of entire static apps (ie: React apps, html pages)
- combine embedded sql database state and file volume mount
- simple GET/POST without multipart-mime

Todo:

- reverse proxy endpoints (A few hundred lines of code at most, as I write reverse proxies often)
- the reverse proxy endpoints would allow full apps to work 
- OpenPolicyAgent for security enforcement, I did this in a separate project, and it took a few hours.

```
# Must be on Linux
./cleanbuild && ./gosqlite
```

Install a react app in a tarball, or a simple html app

```
(
  cd `dirname $0`
  ( cd app && tar cvf ../app.tar . ) 
  curl -X POST --data-binary @app.tar http://localhost:9321/files/app/v1?install=true
  #rm app.tar
)

```

Upload a normal file, one by one

```
  curl -X POST --data-binary @resume.pdf http://localhost:9321/files/doc/rob/resume.pdf
```
Adding reverseproxy endpoints to make full-blown apps work will be easy. Permission system for safe updates a little less so, but not hard.
