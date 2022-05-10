/*
  This pattern MUST be used, because inside of a reverse proxy,
  we count objects by the first part of the path, which is
  assumed to have a low, and meaningful cardinality.

  Here I attempt to just post the file rather than multi-part mime,
  since even though it is the standard for posting files as one transaction,
  it very much baffles users. It's better to just post metadata
  separately from files; which you want to do anyway if
  you are appending to a file over its lifetime.

  You may want to post metadata to override any attributes
  that would infer defaults that you don't want; or to join
  in attributes.  IE: to search for a specific permission
  attribute.

  verb cmd  path
  POST /meta/robf/docs/resume.pdf
        meta - we are posting user metadata about an object
  POST /append/robf/docs/resume.pdf
        append - we are appending content
  POST /eof/robf/docs/resume.pdf
        append - we are done appending content
  POST /files/robf/docs/resume.pdf
        files - same as append+eof
  POST /permission/robf/docs/resume.pdf
        permission - is a language for filtering out content
                   - requires a JWT claim, or user that resolves one

  Perhaps it is best to store files (created in files or append+eof)
  literally on the filesystem; since searchable content needs to
  be stored as a different string anyway.  schema.db and ./files/ would
  be volume mounts.  Permissions and meta may not be stored literally.

  The content is stored literally for being pulled back out.
  The content is also content-searched.  There may need to
  be a transform to preprocess it for suitability for search.

  GET /list/robf/docs
       list - returns an array of hits, in paging metadata
              or as html links, depending on args
  GET /files/robf/docs/resume.pdf
       files - streams of files.  Served straightforwardly
               off of the filesystem with http.FileServer

  GET /permission/robf/docs/resume.pdf
       permission - probably rego language
 */
CREATE TABLE `filemeta` (
	`id` INTEGER PRIMARY KEY AUTOINCREMENT,
	`cmd` TEXT,
	`path` TEXT,
	`name` TEXT,
      `contentType` TEXT,
      `contentSize` INTEGER
);

/*
  GET /search/robf/docs/resume.pdf?q=Rob+Fielding
       search - returns the same format of a listing, of urls that hit
 */
CREATE VIRTUAL TABLE `filesearch` USING FTS5(
	`id`,
      `cmd`,
	`path`,
	`name`,
	`part`,
      `original_path`,
      `original_name`,
	`content`
);
