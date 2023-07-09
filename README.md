# redis-in-action



## 定义article接口

### 第一个接口：PostArticle发布文章

```
生成id,用自增   

初始化：
	自己给文章投票
	设置过期时间，一个星期停止投票，所以不要再维护这个投票人的集合，自动过期老化。
	发布文章到hash  键是"article:"+articleId  
```

### 第二个接口：Vote 给文章投票



### 第三个接口：GetArticles获取文章，排行榜







## 定义ArticleManager实现  + 初始化 

主要是包含redis的连接







