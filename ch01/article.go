package ch01

import (
	"fmt"
	"github.com/go-redis/redis"
	"redisInAction/config"
	"strconv"
	"strings"
	"time"
)

//type User struct {
//	Userid int32
//	Name string
//}

//type Article struct {
//	ArticleId int32
//	Title string
//	Link string
//	poster  User
//	time  time.Time
//	votes int32
//}

type Article interface {
	PostArticle(string,string,string) string
	Vote(string,string)
	GetArticles( int64,string)  []map[string]string

}

type ArticleManager struct {
	Conn *redis.Client
}

func NewArticleManager(conn *redis.Client) *ArticleManager	 {
	return &ArticleManager{Conn: conn}
}

func (r *ArticleManager)PostArticle(user,title,link string) string {
	articleId := strconv.Itoa(int(r.Conn.Incr("article:").Val()))//生成id

	voted:="voted:"+articleId
	r.Conn.Sadd(voted,user)  //投票用户集合，键："voted:"+articleId ，自己给文章投票
	//设置这个集合的过期时间，一个星期停止投票，所以不要再维护这个投票人的集合，自动过期老化。
	r.Conn.Expire(voted,config.ValidTime)

	now:=time.Now().Unix()
	article:="article:"+articleId  //文章编号

	//发布文章  hash  键是"article:"+articleId
	r.Conn.HMSet(article,map[string]interface{}{
		"title":title,
		"link":link,
		"poster":user,
		"time":now,
		"votes":1,
	})
	//加入热度榜集合
	r.Conn.ZAdd("score:", redis.Z{Score: float64(now+config.VoteScore),Member: article})
	//最新榜集合
	r.Conn.ZAdd("time:",redis.Z{
		Score:  float64(now),
		Member: article,
	})

}

func (r *ArticleManager)Vote(article ,user string)  {
	//保护：太早的就不要再投了。
	cutoff:=time.Now().Unix()-config.ValidTime//这个时间之前不要再投了
	if r.Conn.ZScore("time:",article).Val()<float64(cutoff){
		return
	}

	articleId:=strings.Split(article,":")[1]//取出文章号
	if r.Conn.SAdd("voted:"+articleId,user).Val()!=0{//如果没投过，返回不是0，	就是有效投票。不然就是已经在集合里了，算重复投票
		r.Conn.ZIncrBy("score:",config.VoteScore,article)
		r.Conn.HIncrBy(article,"votes",1)
	}


}

func (r *ArticleManager)GetArticles(pages int64 ,order string) []map[string]string {
	//默认返回按热度顺序，万一传进来不是score: 或者time：  。要不要判断order做保护，还是上层已经做好检查了
	if order==""{
		order= "score:"
	}
	//计算开始页数
	start:=(pages-1)*config.ArticlesPerPage//同理：pages如果是异常值呢？上层尽量提前识别和返回错误吧。
	end:=start+config.ArticlesPerPage-1

	ids:=r.Conn.ZRevRange(order,start,end).Val()
	fmt.Println("=======")
	fmt.Println(ids)
	fmt.Println("=======")

	articles:=[]map[string]string{}

	for _,id:=range ids{//拿到这个范围内的文章id  。article:="article:"+articleId
		articleData:=r.Conn.HGetAll(id).Val()//取出这个键对应的hash散列数据。article:="article:"+articleId
		articleData["id"]=id //把id加上后返回。hash不用包含这项，包含的话，不用这步，但每个键对应的hash，需要保存这个额外的id  article:xxxx 键值对
		articles=append(articles,articleData)
	}
	return articles
}

