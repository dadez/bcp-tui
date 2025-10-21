# bcp-tui

This is a project with aim of improve the productivity in my daily work. Thanks to sofa, I have great path to open a link or run a command.
The goal here is mostly to iterate over a list of targets like clusters for open a web link like argocd or run a command like login.

## todo

- [x] read cluster list from a config file, path or remote
- [ ] add util for update cluster list from source
- [x] increase size of last displayed box
- [x] allow run a command on selected clusters, e.g.: `oc get pod -A` or more complex
- [x] preconfigure some commands, e.g.: `oc get postgresCluster -A`
- [ ] esc goes back, do not close
- [x] allow run custom command (prompt)
