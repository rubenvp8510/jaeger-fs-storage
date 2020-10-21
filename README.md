This Jaeger storage plugin stores the spans using the filesystem and retrieve a completed trace using  **traceID**. 

This plugin doesn't have query capabilities, the method `FindTraces(ctx context.Context, query *spanstore.TraceQueryParameters)` ignores the query parameter and return all traces stored.


## Implementation notes

###  Directory Layout

Each trace has his own directory, the name of the directory is the traceID, that directory contains a list of files, each file represents a span, the name of the file is the spanID.

```
└── traces
    ├── <traceID>
    │   ├── <spanID>
    │   ├── <spanID>
    |   .
    |   .
    │   ├── <spanID>
    ├── <traceID>
    │   ├── <spanID>
    │   ├── <spanID>
    ,
    .
```

### Read path

This implementation only supports reading traces by TraceID. For each read request it will list all files under `{basePath}/traces/{traceID}` and unmarshal the spans.

In order to unmarshal the spans, it starts `N` workers and sends the span file path through channel to the workers. 

The workers will consume the channel, open the file and unmarshal it,  then write a result buffer. When all spans are read, the read method returns the results.


### Write path

For write spans to the filesystem when `WriteSpan` is called, it stores the span in a buffered channel.

We have `M` workers watching for the channel and writing the files to the filesystem, the span will be written to: `{basePath}/traces/{traceID}/{spanID}`.
 
The span is marshalled using the proto buffer marshal method before writing to the file.