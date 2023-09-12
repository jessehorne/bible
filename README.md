![bible](bible-logo.png)

---

# Installation

```shell
go install github.com/jessehorne/bible@latest
```

---

# Usage

```shell
> bible --help
Usage: bible [OPTION]...
Access the Holy Bible in your terminal.

  --b=...        Book
                 DEFAULT: "Gen"

  --v=...        Verse(s) (Examples: "1:10-11", "5", "3:16")
                 DEFAULT: Random verse(s)
                  
  --t=...        Version (Examples: "kjv")
                 DEFAULT: "kjv"
                  
  --l=...        Language (Examples: "en")
                 DEFAULT: "en"
  
  -lt            List supported versions.
  --ll        	 List translations for a version.
  --lb           List all books in a version.
  
  -n             Include the number of the verse when printed.

  --help         Show this information.

Examples:
> bible --b="Gen" --v=1:1-2
Genesis 1:1-2
In the beginning God created the heaven and the earth.
And the earth was without form, and void; and darkness was upon the face of the deep. And the Spirit of God moved upon the face of the waters.

> bible --b="Gen" --v=1:1-2 -n
Genesis 1:1-2
1 In the beginning God created the heaven and the earth.
2 And the earth was without form, and void; and darkness was upon the face of the deep. And the Spirit of God moved upon the face of the waters.

For more information, please visit https://github.com/jessehorne/bible
```

---

# Supported Versions & Translations

* **(DEFAULT)** kjv-en via ebible.org (see `data/kjv.db`)
* more coming soon...

---

# License

MIT. See `./LICENSE`.
