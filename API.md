# 社区团购分拣系统 API 文档

## 项目简介

社区团购团长用的后台分拣工具，晚上 10 点截单后自动汇总当天所有订单，按商品分类出总数，按提货点拆分提货单，同时支持供货商对账。后端用 Go 写的，技术栈：Gin + GORM + MySQL。

## 鉴权说明

所有 `/api/*` 开头的接口都需要鉴权，请求头必须带：

```
Authorization: Bearer <token>
```

- token 值由后端环境变量 `AUTH_TOKEN` 控制，部署时会告诉前端具体值
- 没带 Authorization 头 → 返回 401
- 带了但值不对 → 返回 403
- 格式必须是 `Bearer <token>`（中间有空格），少了也返回 401

例外：`GET /ping` 接口不需要鉴权，用来测服务通不通。

## 全局约定

- 所有日期格式：`YYYY-MM-DD`，例如 `2026-06-26`
- 所有金额单位：**元**，不是分，小数两位，例如 `12.50` 表示 12 元 5 角
- 时区统一使用 `Asia/Shanghai`（北京时间）
- 所有接口返回 JSON 格式，结构统一：
  ```json
  { "date": "2026-06-26", "data": ... }
  ```
  出错时：
  ```json
  { "error": "错误信息" }
  ```

## 订单状态枚举

| 值 | 含义 |
|---|---|
| 1 | 待分拣 |
| 2 | 已分拣 |
| 3 | 已提货 |
| 4 | 已取消 |

> 所有汇总/对账/提货单接口**自动排除已取消**的订单，不用前端过滤。

---

## 一、订单模块

### 1.1 创建订单

```
POST /api/orders
```

**请求体：**

```json
{
  "product_id": 1,
  "pickup_point_id": 1,
  "quantity": 2,
  "customer_name": "张三",
  "customer_phone": "13800138000"
}
```

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| product_id | uint64 | 是 | 商品 ID |
| pickup_point_id | uint64 | 是 | 提货点 ID |
| quantity | uint32 | 是 | 购买数量，最小 1 |
| customer_name | string | 否 | 买家姓名 |
| customer_phone | string | 否 | 买家电话 |

**返回（201）：**

```json
{
  "data": {
    "id": 1,
    "order_no": "GB202606262200001234",
    "product_id": 1,
    "pickup_point_id": 1,
    "quantity": 2,
    "unit_price": 5.80,
    "total_price": 11.60,
    "order_date": "2026-06-26",
    "customer_name": "张三",
    "customer_phone": "13800138000",
    "status": 1,
    "created_at": "2026-06-26T20:30:00+08:00",
    "updated_at": "2026-06-26T20:30:00+08:00"
  }
}
```

> `unit_price` 和 `total_price` 由后端自动取商品表的当前单价计算，前端不用传。

### 1.2 查询订单列表

```
GET /api/orders?date=2026-06-26&status=1
```

| 参数 | 必填 | 说明 |
|---|---|---|
| date | 否 | 按日期过滤，不传查全部 |
| status | 否 | 按状态过滤（1/2/3/4），不传查全部 |

**返回（200）：**

```json
{
  "data": [
    {
      "id": 1,
      "order_no": "GB202606262200001234",
      "product_id": 1,
      "pickup_point_id": 1,
      "quantity": 2,
      "unit_price": 5.80,
      "total_price": 11.60,
      "order_date": "2026-06-26",
      "customer_name": "张三",
      "customer_phone": "13800138000",
      "status": 1,
      "product": { "id": 1, "name": "西红柿", "category": "蔬菜", "unit": "斤" },
      "pickup_point": { "id": 1, "name": "阳光小区门口", "address": "..." },
      "created_at": "...",
      "updated_at": "..."
    }
  ]
}
```

### 1.3 修改订单状态

```
PATCH /api/orders/:id/status
```

> `:id` 是路径占位符，换成实际的订单 ID，比如 `/api/orders/1/status`。

**请求体：**

```json
{
  "status": 4
}
```

| 字段 | 类型 | 允许值 | 说明 |
|---|---|---|---|
| status | int8 | 1 / 2 / 3 / 4 | 4 表示取消订单 |

**返回（200）：**

```json
{
  "message": "状态已更新"
}
```

---

## 二、分拣汇总模块

### 2.1 商品汇总

```
GET /api/sorting/summary?date=2026-06-26
```

| 参数 | 必填 | 说明 |
|---|---|---|
| date | 是 | 汇总日期 |

**返回（200）：**

```json
{
  "date": "2026-06-26",
  "data": [
    {
      "product_id": 1,
      "product_name": "西红柿",
      "category": "蔬菜",
      "unit": "斤",
      "total_qty": 45,
      "total_amount": 261.00
    },
    {
      "product_id": 2,
      "product_name": "鸡蛋",
      "category": "禽蛋",
      "unit": "盒",
      "total_qty": 30,
      "total_amount": 300.00
    }
  ]
}
```

> `total_qty` 是当天所有有效订单（排除取消）的该商品总件数。

### 2.2 导出 Excel 提货单

```
GET /api/sorting/export?date=2026-06-26
```

| 参数 | 必填 | 说明 |
|---|---|---|
| date | 否 | 导出日期，不传默认今天 |

**返回：**

直接下载 Excel 文件，文件名格式 `提货单_2026-06-26.xlsx`。

文件结构：
- 每个提货点一个 sheet 页，sheet 名是「提货点名_ID」
- 每个 sheet 里是一个表格，四列：商品名称 | 规格/分类 | 单位 | 件数
- 表头加粗居中，列宽适配中文

---

## 三、提货单模块

### 3.1 按提货点拆分提货单

```
GET /api/pickup/slips?date=2026-06-26
```

| 参数 | 必填 | 说明 |
|---|---|---|
| date | 是 | 查询日期 |

**返回（200）：**

```json
{
  "date": "2026-06-26",
  "data": [
    {
      "pickup_point_id": 1,
      "pickup_point_name": "阳光小区门口",
      "address": "朝阳区阳光路 1 号",
      "grand_total": 561.00,
      "items": [
        {
          "order_id": 1,
          "order_no": "GB202606262200001234",
          "product_name": "西红柿",
          "unit": "斤",
          "quantity": 2,
          "unit_price": 5.80,
          "total_price": 11.60,
          "customer_name": "张三",
          "customer_phone": "13800138000"
        }
      ]
    }
  ]
}
```

> 每个提货点一张单，`grand_total` 是该提货点当天所有订单的总金额。

---

## 四、供货商对账模块

### 4.1 供货商对账（核心接口）

```
GET /api/supplier/reconciliation?date=2026-06-26
```

| 参数 | 必填 | 说明 |
|---|---|---|
| date | 是 | 对账日期 |

**返回（200）：**

```json
{
  "date": "2026-06-26",
  "data": [
    {
      "supplier_id": 1,
      "supplier_name": "老王蔬菜档",
      "stall_number": "A-01",
      "total_amount": 261.00,
      "total_qty": 45,
      "details": [
        {
          "product_id": 1,
          "product_name": "西红柿",
          "quantity": 45,
          "total_amount": 261.00
        }
      ]
    }
  ]
}
```

> **重要**：`total_amount` 的单位是**元**，不是分，保留两位小数。
>
> 只统计**有效订单**（状态 1/2/3），已取消（4）的不会算进来。
>
> 按档口（`stall_number`）分组，每个档口一行，`details` 是该档口供的各个商品明细。

### 4.2 供货商列表

```
GET /api/suppliers
```

**返回：**

```json
{
  "data": [
    {
      "id": 1,
      "name": "老王蔬菜档",
      "stall_number": "A-01",
      "contact": "13800138000"
    }
  ]
}
```

### 4.3 商品列表

```
GET /api/products
```

**返回：**

```json
{
  "data": [
    {
      "id": 1,
      "name": "西红柿",
      "category": "蔬菜",
      "unit": "斤",
      "price": 5.80,
      "supplier_id": 1
    }
  ]
}
```

### 4.4 提货点列表

```
GET /api/pickup-points
```

**返回：**

```json
{
  "data": [
    {
      "id": 1,
      "name": "阳光小区门口",
      "address": "朝阳区阳光路 1 号",
      "contact": "李姐 13800138000"
    }
  ]
}
```

---

## 五、数据库表结构

四张核心表，关联关系：

```
orders
  ├─ product_id → products.id
  └─ pickup_point_id → pickup_points.id

products
  └─ supplier_id → suppliers.id
```

### suppliers（供货商/档口）

| 字段 | 类型 | 说明 |
|---|---|---|
| id | BIGINT UNSIGNED | 主键 |
| name | VARCHAR(100) | 档口名称，例如「老王蔬菜档」 |
| stall_number | VARCHAR(50) | 档口号，例如「A-01」，唯一 |
| contact | VARCHAR(50) | 联系电话 |

### products（商品）

| 字段 | 类型 | 说明 |
|---|---|---|
| id | BIGINT UNSIGNED | 主键 |
| name | VARCHAR(100) | 商品名称 |
| category | VARCHAR(50) | 分类，例如「蔬菜」「禽蛋」 |
| unit | VARCHAR(20) | 计量单位，例如「斤」「盒」「份」 |
| price | DECIMAL(10,2) | 单价（元） |
| supplier_id | BIGINT UNSIGNED | 关联供货商 ID |

### pickup_points（提货点）

| 字段 | 类型 | 说明 |
|---|---|---|
| id | BIGINT UNSIGNED | 主键 |
| name | VARCHAR(100) | 提货点名称 |
| address | VARCHAR(255) | 详细地址 |
| contact | VARCHAR(50) | 联系人/电话 |

### orders（订单）

| 字段 | 类型 | 说明 |
|---|---|---|
| id | BIGINT UNSIGNED | 主键 |
| order_no | VARCHAR(32) | 订单号，唯一 |
| product_id | BIGINT UNSIGNED | 商品 ID |
| pickup_point_id | BIGINT UNSIGNED | 提货点 ID |
| quantity | INT UNSIGNED | 购买数量 |
| unit_price | DECIMAL(10,2) | 下单时的单价（快照） |
| total_price | DECIMAL(10,2) | 小计金额 |
| order_date | DATE | 下单日期，用于按天汇总 |
| customer_name | VARCHAR(50) | 买家姓名 |
| customer_phone | VARCHAR(20) | 买家电话 |
| status | TINYINT | 订单状态：1待分拣 2已分拣 3已提货 4已取消 |

> `unit_price` 存下单时的价格快照，避免后续改商品单价影响历史订单。
