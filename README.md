# Configo 🛶🏠

> A tiny, strongly-consistent configuration store built with Go and HashiCorp Raft.

Configo is a **mini-Consul / mini-etcd** designed to demonstrate real-world distributed systems concepts: **leader election, log replication, fault tolerance, and strong consistency** — all in a compact, understandable codebase.

---

## ✨ Features (MVP)

* **Strongly consistent configuration storage**
* **Leader-based writes** using HashiCorp Raft
* **Replicated state** across all nodes
* **HTTP API** for CRUD operations
* **Persistent storage** (Raft log + state on disk)
* **Automatic leader election & failover**

---

## 📦 API (MVP)

### Put / Update Config

```http
PUT /config/{key}
```

Body:

```json
{
  "value": "some-config-value"
}
```

### Get Config

```http
GET /config/{key}
```

### Delete Config

```http
DELETE /config/{key}
```

### List All Configs

```http
GET /config
```

---

## 🔁 Consistency Model

* **Writes**: Linearizable (via Raft leader)
* **Reads**: Served locally (can be upgraded to leader-only reads)
* **Failure Handling**: Automatic leader re-election

---

## 💾 Persistence

Each node persists:

* Raft log
* Raft snapshots
* Applied configuration state

This allows:

* Safe restarts
* Fast recovery
* Log compaction via snapshots

---

## 🛠 Tech Stack

* **Go**
* **HashiCorp Raft**
* **BoltDB / BadgerDB**
* **net/http**
* **Docker & Docker Compose**
