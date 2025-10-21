# bcp-tui

![GitHub Actions Workflow Status](https://img.shields.io/github/actions/workflow/status/dadez/bcp-tui/go-ossf-slsa3-publish.yml?branch=main)
![GitHub top language](https://img.shields.io/github/languages/top/dadez/bcp-tui)

This is a project with aim of improve the productivity in my daily work. Thanks to [sofa](https://github.com/dadez/sofa) (kudos to
[f4z3r](https://github.com/f4z3r/)), I have great path to open a link or run a command.
The goal here is mostly to iterate over a list of targets like clusters for open a web link like argocd or run a command like login.

## todo

- [x] read cluster list from a config file, path or remote
- [ ] add util for update cluster list from source
- [x] increase size of last displayed box
- [x] allow run a command on selected clusters, e.g.: `oc get pod -A` or more complex
- [x] preconfigure some commands, e.g.: `oc get postgresCluster -A`
- [ ] esc goes back, do not close
- [x] allow run custom command (prompt)
