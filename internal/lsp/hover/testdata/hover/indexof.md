### [indexof](https://www.openpolicyagent.org/docs/latest/policy-reference/#builtin-strings-indexof)

```rego
output := indexof(haystack, needle)
```

Returns the index of a substring contained inside a string.


#### Arguments

- `haystack` string — string to search in
- `needle` string — substring to look for


Returns `output` of type `number`: index of first occurrence, `-1` if not found
