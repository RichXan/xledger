# Xledger AI 记账 Apple Shortcuts 配置指南

## 前提条件

1. 已部署 Xledger 后端（支持 HTTPS）
2. 已有 Personal Access Token（PAT）
3. 有可用的 n8n 实例

## iOS 快捷指令配置

### 方式一：自然语言记账

1. 打开 iOS「快捷指令」App
2. 创建新快捷指令
3. 添加动作：
   - **「文本」**：输入你想记账的内容，如"午饭25元微信"
   - **「URL」**：填入你的 n8n webhook 地址
     ```
     https://your-n8n-domain.com/webhook/quick-entry
     ```
   - **「获取 URL 内容」**：
     - 方法：POST
     - 头部：Content-Type: application/json
     - 请求体：
       ```json
       {
         "text": "午饭25元微信",
         "pat": "你的PAT"
       }
       ```
   - **「显示结果」**：显示上一步的响应内容

4. 为快捷指令设置一个名称和图标
5. 可选：添加 Siri 短语，如"记账"

### 方式二：OCR 截图记账

1. 打开 iOS「快捷指令」App
2. 创建新快捷指令
3. 添加动作：
   - **「截图」** 或 **「从照片选择」**
   - **「识别文本」**（iOS 原生 OCR）
   - **「文本」**：将识别结果存入变量
   - **「URL」**：填入 n8n webhook 地址
   - **「获取 URL 内容」**：POST 方式，请求体同方式一
   - **「显示结果」**

### Siri 触发示例

- "Hey Siri，记账" → 打开快捷指令 → 输入文本

### 测试

1. 在 iOS 快捷指令中运行你的快捷指令
2. 应该收到来自 Xledger 的确认消息
3. 登录 Xledger Web 查看交易是否创建成功
