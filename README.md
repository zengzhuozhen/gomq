<h1>Go-MQ</h1>

implement a sample message queue by golang . :smile:

 :sunny: :sunny: :sunny:
 
 

   
<h3>Quick Start</h3>

    make dockerPrepare
  
    docker-compose up -d 
  
<h3>Command Line</h3>
Publish the message is like that:
    
    docker exec -it gomq gomqctl --topic A --connect 127.0.0.1:9000 pub hello wrold everyone
    
And Subscribe the topic by:
    
    docker exec -it gomq gomqctl --topic A --connect 127.0.0.1:9000 sub 
    
<h3> HTTP </h3>
Also use http is a easier way to do (by example):      
     
    curl http://localhost:8000/messages?topic=A 
     
