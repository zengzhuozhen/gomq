<h1>Go-MQ</h1>

implement a sample message queue by golang . :smile:

 :sunny: :sunny: :sunny:

   
<h3>Quick Start</h3>

  make dockerPrepare
  docker-compose up -d 
  
<h3>Publish</h3>

    docker exec -it gomq gomqctl --topic A --connect 127.0.0.1:9000 pub hello wrold everyone
    
<h3>Subscribe</h3>
    
    docker exec -it gomq gomqctl --topic A --connect 127.0.0.1:9000 sub 
    
