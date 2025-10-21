# bcp-tui

![GitHub Actions Workflow Status](https://img.shields.io/github/actions/workflow/status/dadez/bcp-tui/go-ossf-slsa3-publish.yml?branch=main)
![GitHub top language](https://img.shields.io/github/languages/top/dadez/bcp-tui)

This is a project with aim of improve the productivity in my daily work. Thanks to [sofa](https://github.com/dadez/sofa) (kudos to
[f4z3r](https://github.com/f4z3r/)), I have great path to open a link or run a command.

The goal here is to iterate over a list of targets like clusters for run command, e.g. open an html webpage.

## Configuration

See [config.yaml](./config.yaml) as example.

By default, looking in `$XDG_CONFIG_HOME/bcp-tui`, `$HOME/.config/bcp/tui` and `./` for config files named `config` with multiple format support, see [viper config files](https://github.com/spf13/viper?tab=readme-ov-file#reading-config-files)

You can override the configuration file with the option `-c` with or without file name.

```bash
go run ./main.go -c /tmp/
go run ./main.go -c /tmp/config.yaml
go run ./main.go -c /tmp/config.josn
```

## todo

- [ ] read cluster list from a remote path or url
- [ ] add util for update cluster list from source
- [ ] esc goes back, do not close
- [x] allow run custom command (prompt)
- [ ] add tests
