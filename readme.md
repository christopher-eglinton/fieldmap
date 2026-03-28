# fieldmap

A small Go library for mapping nested external data (`map[string]any`) into strongly-typed structs using declarative rules.

---

## Overview

`fieldmap` helps when working with external data sources (APIs, JSON, CSV) that:

* have different field names
* are nested or inconsistent
* require basic transformations (e.g. trim, lowercase, parsing)

Instead of writing repetitive mapping code, you define rules and apply them.

---

## Features (v1)

* Map nested fields using dot paths (`employee.givenName`)
* Assign values into struct fields by name
* Optional field transformations
* Required field validation
* Aggregated error reporting

---

## Installation

```bash
go get github.com/christopher-eglinton/fieldmap
```

---

## Example

### Input

```json
{
  "employee": {
    "givenName": "Tony",
    "familyName": "Soprano"
  },
  "workEmail": " tone@dimeo.com ",
  "employment": {
    "active": "true",
    "startDate": "2026-03-28"
  }
}
```

---

### Target struct

```go
type Employee struct {
    FirstName string
    LastName  string
    Email     string
    Active    bool
    StartDate time.Time
}
```

---

### Mapping

```go
cfg := fieldmap.Config{
    Rules: []fieldmap.Rule{
        {From: "employee.givenName", To: "FirstName", Required: true},
        {From: "employee.familyName", To: "LastName", Required: true},
        {
            From:      "workEmail",
            To:        "Email",
            Required:  true,
            Transform: fieldmap.TrimLower(),
        },
        {
            From:      "employment.active",
            To:        "Active",
            Transform: fieldmap.StringToBool(),
        },
        {
            From:      "employment.startDate",
            To:        "StartDate",
            Transform: fieldmap.ParseTime("2006-01-02"),
        },
    },
}

var emp Employee
err := fieldmap.Apply(cfg, input, &emp)
if err != nil {
    panic(err)
}
```

---

## Transformations

Built-in helpers:

```go
fieldmap.TrimLower()
fieldmap.StringToBool()
fieldmap.ParseTime("2006-01-02")
```

Custom transform:

```go
Transform: func(v any) (any, error) {
    s := v.(string)
    return strings.ToUpper(s), nil
}
```

---

## Error Handling

Errors are collected per field and returned as a single error.

Example:

```
Email: required field missing from path "workEmail";
StartDate: transform failed: parsing time ...
```

---

## Limitations (v1)

* Input must be `map[string]any`
* Only dot-path access is supported (no arrays)
* Struct fields must be exported
* No automatic JSON decoding (use `json.Unmarshal` first)

---

## Use Cases

* Mapping external API payloads into internal models
* Normalizing inconsistent data structures
* Simple ETL-style transformations
* Pre-processing data before persistence or forwarding

---

## License

GNU General Public License v3.0

