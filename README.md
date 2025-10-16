# bcp-tui

This is a project with aim of improve the productivity in my daily work. We commonly work with a lot of clusters, for facitate common actions like open a browser session for login to argocd, we can use `bcp-tui`.

## todo

- [ ] read cluster list from a config file, path or remote
- [ ] get oc config (username, password) from a config file
- [ ] add some logging
- [ ] increase size of last displayed box
- [ ] allow run a command on selected clusters, e.g.: `oc get pod -A` or more complex
- [ ] preconfigure some commands, e.g.: `oc get postgresCluster -A`
- [ ] esc goes back, do not close
