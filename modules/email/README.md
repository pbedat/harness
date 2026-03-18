# Email Module

Part email client, part email queue. The module combines two concerns: it acts as a **client** for reading, organizing, and composing emails across familiar mailboxes (inbox, drafts, archive), and as a **queue** that controls how messages enter and leave the system — filtering by sender and recipient, enforcing rate limits, and routing deliveries through an event-driven pipeline.

Inbound messages arrive via Postmark webhooks and are staged in the Inbox queue before delivery. Outbound messages move from Draft through the Outbox queue to an external mail provider. All state is persisted to the local filesystem.

## Domain Model

### Mailboxes

Emails live in one of five mailboxes:

| Mailbox   | Purpose                          |
|-----------|----------------------------------|
| `inbox`   | Received emails                  |
| `draft`   | Unsent emails being composed     |
| `outbox`  | Emails queued for sending        |
| `sent`    | Successfully delivered emails    |
| `archive` | Archived emails                  |

### Emails

An email is an immutable aggregate carrying `from`, `to`, `subject`, `body`, `htmlBody`, and `headers`. Once created, only two operations mutate state:

- **MarkAsRead** — flags the email as read
- **Move** — transfers the email to a different mailbox

Both operations produce domain events (`ErstelltEvent`, `MovedEvent`) that drive downstream processing.

### Queues

Inbox and Outbox mailboxes each have an associated queue that acts as a staging area. Queues support filtering by allowed recipients and sender, and enforce per-queue limits. Enqueue/dequeue operations produce events that trigger delivery handlers.

## Data Flow

### Inbound (Postmark → Inbox)

```
Postmark webhook
  → POST /webhooks/postmark/inbound (basic auth)
  → Enqueue command (Inbox queue)
  → EnqueuedEvent
  → InboxDeliveryHandler dequeues and creates Email
  → Email stored in inbox/
```

### Outbound (Draft → Sent)

```
Draft email created
  → Move command (Draft → Outbox)
  → MovedEvent
  → OutboxDeliveryHandler queues to Sent queue
  → EnqueuedEvent
  → Dequeue, send via mail port, create Email in Sent
  → Email stored in sent/
```
