# jsongostruct

generate golang struct from json schema

## example

```
cat /path/to/jsonfile | jsongostruct
```

```json
{
  "url": "http://example.com",
  "id": 12345,
  "name": "web",
  "bool": true,
  "array": [
    "foo",
    "bar"
  ],
  "map": {
    "foo": "bar",
    "dameleon": "dame"
  }
}
```

```go
type XXX struct {
        Url   string   `json:"url"`
        Id    float64  `json:"id"`
        Name  string   `json:"name"`
        Bool  bool     `json:"bool"`
        Array []string `json:"array"`
        Map   struct {
                Foo      string `json:"foo"`
                Dameleon string `json:"dameleon"`
        } `json:"map"`
}
```

## TODO

* testing
* tag position for struct type
* nil value handling ( current, null value's type is &lt;nil&gt; )
