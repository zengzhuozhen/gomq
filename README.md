<h1>Go-MQ</h1>

implement a sample message queue by golang . :smile:


ALREADY IMPLEMENT
 -
 - MQTT 协议
 - Topic	

TODO 
 -
 - Feature
    - partition
    - data store
    - transaction message
    - 
    
 - Fix
    - ~~Bug : 多个消费者消费消息貌似出现拿错数据的问题~~ 
    - ~~Bug : 貌似消费端会有所谓的 "粘包" 的问题，也许要定义一个通讯协议
   处理这个问题 EOF \r\n ?~~
   - ~~经过一次发布消费流程后退出消费者会导致server panic ，等待修复~~
 
 :sunny: :sunny: :sunny:
    

