# 数据模型设计

## 网关分组`node_group`

* 表结构 

|key| type|desc|
|---|---|---| 
|name|string|分组名称，需要保障唯一|
|desc|string|分组描述|

* 说明  
   提供网关分组 在保障envoy节点名称唯一的基础上进行分组，保障每个分组绑定的监听器及路由信息统一

## 分组节点绑定`node_group_bind`

|key| type|desc|
|---|---|---| 
|group_name|string|分组名称，需要保障唯一|
|node_id|string|envoy 实例id|

## 监听配置(listeners)
|key| type|desc|
|---|---|---| 
|name|string|监听名称|
|address|string|绑定地址，默认0.0.0.0|
|port|int32|绑定端口号|
|metadata|json|元数据信息|
|group_name|string|绑定envoy 分组|


## 过滤器模板`filter_template`

|key| type|desc|
|---|---|---| 
|name|string|envoy 默认的filter 名称|
|template|json|json模板|
|metadata|json|元数据信息|

## 过滤器监听绑定`listener_bind_filter`

|key| type|desc|
|---|---|---| 
|listener_name|string|监听名称|
|filter_name|string|envoy 默认的filter 名称|
|filter_config|json|通过template转换成的配置|

## 应用集群`cluster`

|key| type|desc|
|---|---|---| 
|name|string|监听名称|
|endpoint_discovery|string|后端发现配置名称|
|desc|string|描述|

## 应用服务发现`endpoint_discovery_config`

|key| type|desc|
|---|---|---| 
|name|string|配置名称|
|server_address|string|服务发现address 和端口号，同配置需要进行不同的解析方法|
|type|string|类型zookeeper,consul,kubernetes|
|auth|string|认证信息|

## 应用endpoint `endpoints`

|key| type|desc|
|---|---|---| 
|cluster_name|string|应用集群名称|
|address|string|服务地址|
|port|int32|端口号|

## 路由配置   
|key| type|desc|
|---|---|---| 
|name |string|路由名称|
|listener_name|string|绑定监听器|


## 路由过滤器绑定

|key| type|desc|
|---|---|---| 
|id |uuid|绑定id|
|router_name|string|路由名称|
|filter_name|string|envoy 默认的filter 名称|
|filter_config|json|通过template转换成的配置|
 

 ## 路由cluster绑定

|key| type|desc|
|---|---|---| 
|id |uuid|绑定id|
|router_name|string|路由名称|
|cluster_name|string|envoy 默认的filter 名称|
|path|string|转发路径|
|path_type|string|转发路径类型 `Prefix` `Path` `Regex`|
|header|json|header匹配转发，key value and type 类型的,type 值`Contains` `SuffixMatch` `PrefixMatch` `PresentMatch` `RangeMatch` `SafeRegexMatch` `ExactMatch` default `ExactMatch`|


 
 
 