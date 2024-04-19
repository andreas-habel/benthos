### Metadata

This output supports different metadata headers to set various message headers and properties.

#### `solace_priority`

Set the metadata to a number between `9` (highest) and `0` (lowest) priority to use message priority.
To have messages sorted in queues based on the priority, make sure that `Respect Message Priority` is set to `true` for the queue.
When enabled, messages contained in the Queue are delivered in priority order.
Regardless of this setting, message priority is not respected when browsing the queue, when the queue is used by a bridge, or if the queue is partitioned.
