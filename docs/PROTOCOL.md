# Protocol v1

Payload (PS5) listens on TCP `9045`. Each connection receives one JSON line, then is closed.

```json
{"v":1,"running":true,"title_id":"PPSA01234","title_name":"Game Name"}
```

| field        | meaning                                              |
|--------------|------------------------------------------------------|
| `v`          | protocol version                                     |
| `running`    | `false` on the home screen / no game                 |
| `title_id`   | PPSA/CUSA id; companion detects game changes by this |
| `title_name` | display name (may be empty -> id is shown)           |

Stateless on purpose: the PC polls, so the payload needs no Discord code
and survives the PC/Discord restarting.
