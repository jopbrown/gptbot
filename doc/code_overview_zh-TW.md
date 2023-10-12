# Code Analysis

該軟體是一個聊天機器人系統，能夠與使用者進行自然語言對話，並提供多種對話角色和功能擴展機制，以實現更加靈活的應用場景。

| File | Description |
| --- | --- |
| cmd/gptbot/main.go | 執行gptbot的主程式，負責接收來自用戶的訊息並回應。主要包含初始化機器人的設定、連接聊天平台、處理訊息、調用gpt模型產生回應、發送回應等步驟。 |
| pkg/cfgs/config.go | 讀取、寫入gptbot的配置信息。包括讀取聊天平台的配置、gpt模型的配置、角色配置等，以及在程序運行時允許動態更改這些配置。 |
| pkg/cfgs/role.go | 定義了機器人在聊天群組中所擔任的角色及其權限。可以設置哪些用戶可以對機器人進行控制、設置用戶的權限等。支持將角色信息寫入文件中以便持久化存儲。 |
| pkg/chatbot/bot.go | 定義了一個ChatBot結構，作為整個聊天機器人系統的入口，並提供初始化和運行機器人等方法。ChatBot結構包含設定信息、會話管理器、接口用戶、機器人狀態等。 |
| pkg/chatbot/line.go | 實現了與Line聊天機器人API的交互，並處理Line聊天事件的回調。當有新的Line事件到達時，Bot將事件傳遞給不同的Task進行處理，包括ChatTask、ClearSessionTask、ChangeRoleTask等。 |
| pkg/chatbot/session.go | 定義了會話管理器，並實現了與不同會話存儲介質的交互，如Redis和內存。會話管理器用於管理每個用戶的會話狀態，包括用戶的ID、當前會話ID、會話開始時間、當前角色、是否是首次對話等信息，並提供許多方法來訪問和更改這些信息。 |
| pkg/chatbot/task.go | 定義了一個 `Task` 結構體，它表示機器人所執行的任務。 `Task` 包含一個方法，用於處理從用戶端收到的消息。該方法將消息作為參數，然後根據消息的類型和內容決定該執行哪個任務。 `Task` 還包括一個方法，用於創建回覆消息，以向用戶端發送回覆。這個檔案提供了一個框架，可以根據需要添加更多的任務，以擴展聊天機器人的功能。 |
| pkg/chatbot/bot.go | 實作了機器人的主要邏輯，包括對外API的接口實現和調用OpenAI的API，處理使用者輸入的文本，並組織回應的文本和訊息 |
| pkg/chatbot/line.go | 處理與Line平台的對接，包括對Line平台的Webhook進行驗證和回覆，處理Line使用者發送的訊息 |
| pkg/chatbot/session.go | 管理使用者和機器人的對話Session，維護Session的狀態，如Session的過期和清除歷史訊息等 |


## cmd/gptbot/main.go

概述：
- 该文件是 GPTBot 项目的主要入口文件 `main.go`。
- 它引用了 `cmd/gptbot` 下的其他包和 `github.com/jopbrown` 的一些包。
- `main` 函数负责解析配置，初始化日志和聊天机器人，以及启动服务。
- `run` 函数是 `main` 函数的子函数，它执行了具体的初始化和启动服务的操作。
- `applyLog` 函数是 `run` 函数的子函数，用于初始化日志。

代码逻辑：
- `main` 函数负责解析配置文件并初始化聊天机器人和日志，并启动聊天机器人服务。
- `run` 函数解析配置文件，初始化聊天机器人和日志，最后启动聊天机器人服务。
- `applyLog` 函数用于初始化日志，它会将日志输出到文件和控制台。

其中，程序启动时会打印出 `BuildName`, `BuildVersion`, `BuildHash`, `BuildTime` 四个变量的值，分别表示构建名称、版本、哈希值和构建时间。同时，程序会加载配置文件 `gptbot.yaml` 并覆盖默认配置。聊天机器人服务启动后，会持续运行，直到程序退出。

## pkg/cfgs/config.go

檔案名稱：pkg/cfgs/config.go

這個檔案定義了一個叫做 Config 的結構，它包含了這個應用程式需要的許多設定。這些設定包括：

- DebugMode：bool 型態，代表是否在除錯模式下運行。
- LineChannelToken：string 型態，代表 Line 頻道的 token。
- LineChannelSecret：string 型態，代表 Line 頻道的 secret。
- ChatGptApiUrl：string 型態，代表與 ChatGPT 的 API 通訊 URL。
- ChatGptAccessToken：string 型態，代表與 ChatGPT 通訊的存取權杖。
- SessionExpirePeriod：time.Duration 型態，代表 Session 的過期時間。
- SessionClearInterval：time.Duration 型態，代表 Session 清除的間隔時間。
- DefaultRole：string 型態，代表預設角色。
- Roles：Roles 型態，代表所有可用的角色。
- ServePort：int 型態，代表服務器運行的端口號。
- MaxTaskQueueCap：int 型態，代表最大的任務佇列容量。
- LogPath：string 型態，代表日誌文件的儲存路徑。
- CmdsTalkToAI：[]string 型態，代表與 AI 通訊的命令。
- CmdsClearSession：[]string 型態，代表清除 Session 的命令。
- CmdsChangeRole：[]string 型態，代表更改角色的命令。

此外，這個檔案還定義了幾個用來讀取、儲存、合併 Config 的方法：

- DefaultConfig：用來獲取默認的 Config。
- LoadConfig：用來從文件中讀取 Config。
- ReadConfig：用來從 io.Reader 中讀取 Config。
- Merge：用來將兩個 Config 合併。
- MergeDefault：用來將當前的 Config 與默認的 Config 合併。
- SaveConfig：用來將 Config 儲存到文件中。
- WriteConfig：用來將 Config 寫入 io.Writer 中。

## pkg/cfgs/role.go

檔案名稱: pkg/cfgs/role.go

此程式碼檔案主要負責對聊天機器人中的角色進行操作，包括讀取、寫入、儲存、加載角色。

此檔案包含以下幾個功能:
- 常數 ROLE_GROUP_REBOT 被定義為一段預設值。
- 類型 Roles 被定義為 map[string]string。
- 函式 DefaultRoles() 會讀取 default/roles.yaml 的資料，將它解析成 Roles 類型，然後返回 Roles 類型的變數 roles。
- 函式 LoadRoles(fname string) (Roles, error) 會讀取 fname 檔案的資料，將它解析成 Roles 類型，然後返回 Roles 類型的變數 roles，如果發生錯誤，則返回 error 類型的變數。
- 函式 ReadRoles(r io.Reader) (Roles, error) 會讀取 io.Reader 類型的 r 變數，將它解析成 Roles 類型，然後返回 Roles 類型的變數 roles，如果發生錯誤，則返回 error 類型的變數。
- 函式 SaveRoles(fname string) error 會將 Roles 類型的變數 roles 儲存到 fname 檔案，如果發生錯誤，則返回 error 類型的變數。
- 函式 WriteRoles(w io.Writer) error 會將 Roles 類型的變數 roles 寫入到 io.Writer 類型的 w 變數，如果發生錯誤，則返回 error 類型的變數。

## pkg/chatbot/bot.go

文件名：pkg/chatbot/bot.go

這個文件定義了一個名為 Bot 的結構體，包含以下方法：

- NewBot: 初始化 Bot 實例
- UpdateApiServerAccessToken: 發送 HTTP PATCH 請求更新 API Server Token
- Serve: 啟動 HTTP Server，並同時啟動清理 Session 和 Task 隊列的 goroutine
- DoTasks: 從 Task 隊列讀取 Task 並執行
- ClearExpiredSessionsPeriodically: 定期清理過期的 Session
- Stop: 關閉 Bot 實例，停止所有 goroutine

Bot 結構體包含以下屬性：

- cfg: 設定檔實例
- lineClient: LineBot 客戶端
- gptClient: GPT3 客戶端
- sessMgr: Session 管理器
- taskQueue: Task 隊列
- handler: Gin 引擎實例
- stop: 用來通知所有 goroutine 停止運作的 channel
- userNameCache: 用戶名快取，為一個 map[string]string

Bot 實例提供 HTTP 回撥方法，包括 pingHandler 和 stopHandler，以及 LineBot 回撥方法 linebotCallback，皆為 Gin 路由方法。其餘方法均為私有方法，並不對外開放。

## pkg/chatbot/line.go

這個檔案是聊天機器人(LineBot)的回呼函數。以下是程式碼的邏輯：

- linebotCallback：對於收到的每個事件，檢查它是否為消息事件。如果是，則根據消息的類型進行以下操作：
  - TextMessage：獲取消息內容並清除首尾空格。如果是清除會話命令（bot.cfg.CmdsClearSession中所列命令之一），將一個ClearSessionTask放入Bot.taskQueue佇列中，進行會話清除操作。否則，如果是更改權限命令（bot.cfg.CmdsChangeRole中所列命令之一），則將一個ChangeRoleTask放入Bot.taskQueue佇列中，進行權限更改操作。如果都不是上述兩種情況，則判斷消息是否是發給機器人的命令或者來自私聊消息。如果是，則根據userID獲取使用者名稱，並將一個ChatTask放入Bot.taskQueue佇列中，進行聊天操作。
  - 其他消息類型：忽略
- lineReplyFnWithToken：生成一個回復函數，它會將回復和圖片URL作為LineBot消息發送出去。
- lineGetUserName：根據userID獲取使用者名稱，如果無法獲取則返回一個預設值“路人甲”。
- lineIsGroupEvent：判斷是否為群聊事件。
- lineGetSessionID：根據事件的來源類型，獲取會話ID（用於在會話中保持狀態）。

以上是“pkg/chatbot/line.go”檔案中程式碼的邏輯。

## pkg/chatbot/task.go

以下是對該代碼文件的概述：

- 包chatbot是整個程序的核心。它包含了所有與聊天機器人的業務邏輯相關的代碼。
- Task接口定義了聊天機器人執行的任務，它有一個Do函數，用於執行該任務。
- ChatTask結構體表示聊天任務，它包含了聊天所需的所有參數。當聊天任務被執行時，它會向GPT-3.5模型發送一個消息，該消息包含了當前會話的所有內容，然後模型會返回一個AI助手的回答。聊天過程中所有的消息都會被記錄在日誌文件中。
- ClearSessionTask結構體用於清空當前會話的內容。當該任務被執行時，所有與當前會話相關的內容都會被清空，包括聊天記錄、會話ID等等。
- ChangeRoleTask結構體用於改變當前會話的角色。當該任務被執行時，會話的角色將被修改為指定的角色，並且會話ID也會被重新生成。

## pkg/chatbot/session.go

`pkg/chatbot/session.go` 這個檔案定義了兩個類別，`SessionManager` 和 `Session`，以及一些與之相關的方法，這兩個類別提供了一個會話（session）管理的機制。

`SessionManager` 類別有以下屬性：

- `DefaultRole`：預設的身份。
- `Sessions`：以字串為索引的會話列表，以 `Session` 為值。

`NewSessionManager` 方法會回傳一個新的 `SessionManager` 物件，並將預設身份和空的會話列表賦值。

`GetSession` 方法會回傳指定 ID 的會話。如果找不到相對應的會話，則會建立一個新的會話，並將它加入 `Sessions` 列表中。

`ClearExpiredSessions` 方法會將超過指定過期時間的會話清除，回傳被清除的會話 ID 列表。

`Session` 類別有以下屬性：

- `ID`：會話 ID。
- `Role`：會話身份。
- `Messages`：與該會話相關的聊天記錄，以 `openai.ChatCompletionMessage` 為值。
- `LastUpdateDate`：最後一次更新的時間。

`NewSession` 方法會回傳一個新的 `Session` 物件，並將 `Messages` 設為空陣列，`LastUpdateDate` 設為現在時間，以及 `ID` 和 `Role` 分別設為傳入的值。

`Clear` 方法會清除 `Messages`，並將 `LastUpdateDate` 設為現在時間。

`ShortID` 方法會回傳該會話的短 ID，如果 ID 長度小於等於 12 則直接回傳，否則回傳前 12 個字元。

`AddMessage` 方法會將一條新的聊天記錄加入 `Messages`，並將 `LastUpdateDate` 設為現在時間。

`ChangeRole` 方法會先呼叫 `Clear` 方法清除 `Messages`，然後將 `Role` 設為傳入的值。
