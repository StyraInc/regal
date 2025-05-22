### [json.filter](https://www.openpolicyagent.org/docs/policy-reference/#builtin-object-jsonfilter)

```rego
filtered := json.filter(object, paths)
```

Filters the object. For example: `json.filter({"a": {"b": "x", "c": "y"}}, ["a/b"])` will result in `{"a": {"b": "x"}}`). Paths are not filtered in-order and are deduplicated before being evaluated.


#### Arguments

- `object` object[any: any] — object to filter
- `paths` any<array[any<string, array[any]>], set[any<string, array[any]>]> — JSON string paths


Returns `filtered` of type `any`: remaining data from `object` with only keys specified in `paths`
