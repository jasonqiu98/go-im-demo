# Go IM Demo

A simple terminal IM in Golang as a refresher of Golang syntax, borrowed from the [video](https://www.bilibili.com/video/BV1gf4y1r79E?p=37) and the [docs](https://www.yuque.com/aceld/mo95lb/dsk886) (content in Chinese).

## Quickstart

1. Use `make build` to build the client and the server. Then choose either one of the two following options to get a taste of this little demo.
2. [Option 1] Start the server by `./builds/server`, and then run `nc localhost 8888` to start several clients. Then send the messages in one of the following formats.
   - `{rename}...` to update username to a new one, e.g., `{rename}Alice` to update username to Alice
   - `who` to see who is online
   - `{to=...}...` to send a DM to another user, e.g., `{to=Alice}Hello` to send the direct message "Hello" to Alice
   - any other input for a group chat / broadcasting mode
3. [Option 2] Start the server by `./builds/server` and then start several clients by `./builds/client`. Then follow the instructions of the REPL.

## Some facts

- Any user will be kicked out of the chat room if they don't talk for more than five minutes.
- This is a super micro demo. No config file is involved, but several configuration terms are hardcoded in the files like the following ones...
  - timeout for an inactive user: 5 mins
  - ip and port of the server: "localhost:8888" or "127.0.0.1:8888"
  - ip and port that a client should dial, also "localhost:8888", but this can be changed by the command line options. For example, use `./builds/client -ip 127.0.0.1 -port 8000` to dial "127.0.0.1:8000"
- TCP connections are built with the help of the standard Go library "net".