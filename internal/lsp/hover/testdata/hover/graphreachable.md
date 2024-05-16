### [graph.reachable](https://www.openpolicyagent.org/docs/latest/policy-reference/#builtin-graph-graphreachable)

```rego
output := graph.reachable(graph, initial)
```

Computes the set of reachable nodes in the graph from a set of starting nodes.


#### Arguments

| Name      | Type                                   | Description                                              |
|-----------|----------------------------------------|----------------------------------------------------------|
| `graph`   | object[any: any<array[any], set[any]>] | object containing a set or array of neighboring vertices |
| `initial` | any<array[any], set[any]>              | set or array of root vertices                            |


Returns `output` of type `set[any]`: set of vertices reachable from the `initial` vertices in the directed `graph`
