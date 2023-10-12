# gptbot

gptbot is a Line group chat assistant robot, powered by chatgpt with powerful AI.

## Install

```go
go install github.com/jopbrown/gptbot/cmd/gptbot@latest
```

## Features

* The robot can send private message or join chat groups.
* In groups, it can identify messages from different people.
* You can assign different roles to the robot.
* It supports displaying simple images.

> To interact with AI in a chat group, must begin your message with '@ai'.

## Commands in chat

* Talk to AI in a chat group
    * `@ai`
    * `@小愛`
    * `@小爱`

* Clear chat session
    * `/refresh`
    * `/clear`
    * `/clean`
    * `/清空`

* Switch role
    * `/cosplay <role>`
    * `/扮演 <role>`

* Display all available roles
    * `/cosplay`
    * `/扮演`

## Config

The configuration file must be named `gptbot.yaml` and located next to the executable.

```yaml
LineChannelToken: xxxxxxxxx # line channel token
LineChannelSecret: xxxxxxxxxx # line channel secret
ChatGptAccessToken: xxxxxxxxxx # openai token
ServePort: 8888
```

> To retrieve environment variables, use the following format: `${env.VARNAME}`.

Add more custom roles

```yaml
Roles:
  Role 1: Role Prompt
  Role 2: Role Prompt
```

Using an unofficial OpenAI API-compatible  service.

```yaml
ChatGptApiUrl: http://your_service_url
# default is https://api.openai.com/v1
```

## Development document

Chinese document generated by [codesum](https://github.com/jopbrown/codesum).
- [正體中文](./doc/code_overview_zh-TW.md)
