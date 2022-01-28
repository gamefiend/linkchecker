## Linkchecker

## ideas 

- use terminal coloring to make it easier to read
    - broken links in red
    - working links in green
    - (color blind mode?)
- [ ] redirects 
- [x ]  default output json

## TODO
- [ ] Warning Mode
- [ ] Streaming Filter
- [ ] Redirect
- [ ] Test Flag Logic


### Default
```
linkchecker --url https://mysite.com/quinn.html
:: There are 5 broken links, 10 working links, 3 redirects::
-- Broken --
(404) https://broken1.com/blah
(404) https://broken2.com/blah
(500) https://broken3.com/blah
(403) https://broken4.com/blah
(404) https://broken5.com/blah
```

### Filter status
```
linkchecker --url https://mysite.com/quinn.html
:: There are 5 broken links, 10 working links::
-- Broken --
(404) https://broken1.com/blah
(404) https://broken2.com/blah
(500) https://broken3.com/blah
(403) https://broken4.com/blah
(404) https://broken5.com/blah
```

### Verbose mode
```
linkchecker --url https://mysite.com/quinn.html --verbose
:: There are 5 broken links, 10 working links::
-- Broken --
(404) https://broken1.com/blah
(404) https://broken2.com/blah
(500) https://broken3.com/blah
(403) https://broken4.com/blah
(404) https://broken5.com/blah
-- Working --
https://working1.com/site
....
```
---
pseudocode

linkcheck = CheckPageForLinks(page)

checkpagelinks(page)
    fetch the page 
    record that it's checked
    if down, report
    if up and internal
        grab links 
        for each link
            skip if checked already
           checkpagelinks(page) 

        

/home
    - /tutorials
        - /maps
            - /home
            - /tutorials/pointers
    - youtube.com
        - 1 million cat videos
