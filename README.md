# Web Monitor
This is a simple Go program used to monitor web pages (such as job board like Linkedin or Seek)

### Configuration

Program configuration is stored in `config.json`

- `checkInterval` - Time between each check, in minutes
- `sources` - web pages to check
  - `name` - source name
  - `active` - boolean, set to false to disable source
  - `url` - URL of page to be checked
  - `itemPath` - css selector for root container of a item (e.g. a job)
  - `titlePath` - css selector for item title. It will be prepended by `itemPath` css selector
  - `footerPath` - css selector for page footer, program will wait for footer to be visible before processing
  - `contents` - additional attributes to be displayed
    - `name` - name of the attribute
    - `path` - css selector for the attribute (e.g. job location). It will be prepended by `itemPath` css selector
    - `type` - type of the attributes. Available values:
      - `text` - Will print out the text inside HTML element directly
      - `list` - `path` may select multiple HTML elements. Each element will be printed out in a new line
      - `list-inline` - same as `list` except all list items will be printed in one line
      - `url` - Will get the `href` attribute's value of the HTML element

#### Reference

During developing these resources provide great information.

- https://itnext.io/scrape-the-web-faster-in-go-with-chromedp-c94e43f116ce
- https://gregtczap.com/blog/golang-scraping-strategies-frameworks/
- https://pkg.go.dev/github.com/chromedp/chromedp#WaitReady
