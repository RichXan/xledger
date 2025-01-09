# Prometheus

## 简介

Prometheus 通过 HTTP 协议从远程机器上的exporter收集数据，并将其存储在本地磁盘上的时间序列数据库中。用户可以通过 PromQL 查询数据，并使用 Grafana 进行可视化。

Prometheus 的架构设计使其具有高可用性和可扩展性，可以轻松地部署在多个服务器上，以实现大规模的监控。

prometheus 是一个开源的监控系统，用于收集、存储和查询时间序列数据。它由多个组件组成，包括：

- Prometheus Server：用于收集和存储时间序列数据。
- Node Exporter：用于收集主机级别的监控数据。
- Exporter：用于收集特定应用程序的监控数据。(监控适配器：redis, mysql, mongodb, kafka, etc)
- Alertmanager：用于处理告警。
- Grafana：用于可视化监控数据。

## 支持类型
Prometheus为了支持各种中间件以及第三方的监控提供了exporter，大家可以把它理解成监控适配器，将不同指标类型和格式的数据统一转化为Prometheus能够识别的指标类型。

例如Node exporter主要通过读取Linux的/proc以及/sys目录下的系统文件获取操作系统运行状态，reids exporter通过Reids命令行获取指标，mysql exporter通过读取数据库监控表获取MySQL的性能数据。他们将这些异构的数据转化为标准的Prometheus格式，并提供HTTP查询接口。

- 主机监控
- 容器监控
- 中间件监控
- 数据库监控
- 网络监控
- 日志监控
- 应用监控
- 自定义监控
