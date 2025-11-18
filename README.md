# Goon

**Goon** is a Go library for **serializing Go data** structures into [Token-Oriented Object Notation (TOON)](https://github.com/toon-format/toon), a compact, human-readable, schema-aware format designed to minimize token usage for large language models (LLMs).

TOON provides a lossless representation of JSON objects, arrays, and primitives while being more token-efficient for structured data, especially uniform arrays of objects. Goon allows Go developers to encode virtually **any Go type**—structs, pointers, slices, arrays, maps, and primitive types—into TOON.

---

# Features
- Serialize all Go native types to TOON (structs, maps, slices, arrays, primitives)
- Efficient token usage, ideal for LLM prompts
- Supports nested objects and mixed arrays
- Lossless conversion for JSON-compatible Go data

> ⚠️ Currently, Goon only supports serialization from Go to TOON. Deserialization (TOON → Go) is not yet implemented.

---

## Installation

```bash
go get github.com/roboogg133/goon
```


# Usage Examples

```go
package main

import (
    "fmt"
    "log"

    "github.com/roboogg133/goon/goon"
)

type User struct {
    ID     int     `toon:"id"`
    Name   *string `toon:"name"`
    Active bool    `toon:"active"`
    Email  string  `toon:"email"`
    Score  float32 `toon:"score"`
}

func main() {
    name := "Ada Lovelace"
    user := User{
        ID:     123,
        Name:   &name,
        Active: true,
        Email:  "ada@example.com",
        Score:  98.5,
    }
    
    data, err := goon.Marshal(user)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println(string(data))
}
```
### Result:
```toon
id : 123
name : Ada Lovelace
active : true
email : ada@example.com
score : 98.5
```

---

## Nested Structures

```go
type Nested struct {
    User struct {
        ID      int    `toon:"id"`
        Name    string `toon:"name"`
        Content struct {
            Email string `toon:"email"`
            Phone string `toon:"phone"`
        } `toon:"contact"`
    } `toon:"user"`
    
// Work with maps too!
nestedMap := map[string]map[string]any{
			"user": {
				"id":   123,
				"name": "Ada Lovelace",
				"contact": map[string]string{
					"email": "ada@example.com",
					"phone": "+1-555-0100",
				},
				"settings": map[string]any{
					"theme":         "dark",
					"notifications": true,
				},
			},
		}   
```
### Result:
```toon
user :
  id : 123
  name : Ada Lovelace
  contact :
    email : ada@example.com
    phone : "+1-555-0100"
  settings :
    theme : dark
    notifications : true
```

---

## Arrays
```go
type Arrays struct {
    Tags    []string `toon:"tags"`
    Numbers []int    `toon:"numbers"`
    Empty   []string `toon:"empty"`
}
```
### Result:
```toon
tags[3]: admin,ops,dev
numbers[5]: 1,2,3,4,5
empty[0]:
```

---

## Mixed Arrays
```go
arrays := map[string][]any{
			"tags":    {"admin", "ops", "dev"},
			"numbers": {1, 2, 3, 4, 5},
			"empty":   {},
}
```
### Result:
```toon
tags[3]:
  - admin
  - ops
  - dev
numbers[5]:
  - 1
  - 2
  - 3
  - 4
  - 5
empty[0]:
```

---

### Raw slice
```go
[]string{"a", "aa", "bbb", "ccc", "dddd", "true", " padding "}
```
### Result:
```toon
[7]: a,aa,bbb,ccc,dddd,"true"," padding "
```

Goon efficiently serializes all these Go types to TOON, producing human-readable output suitable for LLMs, logging, or configuration files.

---

# Why TOON?
TOON is a compact, readable format for structured data, combining:
- **CSV-like layout** for uniform arrays
- Optimized for token usage in AI prompts
It is particularly useful for **large arrays of objects**1, while maintaining full JSON compatibility.
---
## Example Test Cases
Modernized tests demonstrate serialization of:
- Simple structs with pointers and primitives
- Nested structs/maps
- Mixed-type arrays
- Arrays of primitives
- Slices of maps/structs



# TODO
- Implement deserialization (TOON → Go)

# References
[Toon Specification](https://github.com/toon-format/toon)
