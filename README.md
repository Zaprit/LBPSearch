# LBPSearch

This is a search engine for the LittleBigPlanet server archive that someone published on archive.org.
It supports viewing levels, viewing users, downloading levels, and adding them to be played on refresh.

It's written in go with speed in mind as nobody wants to be stuck waiting for a search to complete.

## Configuration

You need only one file, lbpsearch.yaml for this to work. Here's an example lbpsearch.yaml:

```yaml
# location of the cache, converted icons and level archives
cache_path: /path/to/cache/lbpsearch_cache
# global url of the server
global_url: https://zaprit.fish
# where the archive zips are, currently unused
archive_path: /path/to/dry/archive

# path to the archive_dl command required for level downloading, source can be found here https://github.com/Zaprit/lbp_archive_dl
archive_dl_command_path: /home/henry/bin/archive_dl

# Database connection settings, it's a postgres DB
database_name: lbpsearch
database_user: lbpsearch
database_password: lbpsearch
database_host: localhost
database_port: 5432
database_sslmode: disable

# Show a github sponsors message for Zaprit on search results (if you're self hosting you probably don't want this, idk)
show_sponsor_message: true

# Add something to the header of the page, for things like analytics
header_injection: |
  <script src="https://some-analytics.com/idk" />
```
