Preinstall

Edit /etc/security/limits.conf 

		root soft nofile 65536
		root hard nofile 65536
		* soft nofile 65536
		* hard nofile 65536

Reboot

Ruby and Fluentd

		gpg --keyserver hkp://keys.gnupg.net --recv-keys 409B6B1796C275462A1703113804BB82D39DC0E3
		
		curl -sSL https://get.rvm.io | bash -s stable
		
		source /home/vagrant/.rvm/scripts/rvm
		rvm install 2.2.0
		rvm --default use 2.2.0
		gem install fluentd --no-ri --no-rdoc
		sudo apt-get install libcurl3 libcurl3-gnutls libcurl4-openssl-dev
		gem install fluent-plugin-elasticsearch

Write of json data

Note - need to make sure there are no proxies set to enable connectivity to elasticsearch.
Also note the time_key is a 'fake' json field as fluentd cannot parse the
RFC 3339 formatted timestamp,

		<source>
		  type tail
		  path /home/vagrant/xavi-logs/demo.log
		  pos_file /home//vagrant/xavi-logs/demo-log-pos
		  tag foobar.json 
		  format json 
		  time_key fake
		</source>
				
		<match **>
		  type copy
		  <store>
		  type elasticsearch
		  logstash_format true
		  host 172.20.20.30
		  port 9200
		  flush_interval 10s
		  </store>
		</match>

		
To get around the proxy issue, I started fluentd using the following script
<pre>
#!/bin/sh	
unset http_proxy
unset https_proxy
unset HTTP_PROXY
unset HTTPS_PROXY
curl 172.20.20.30:9200
fluentd -v -c fluent-xavi.conf
</pre>

May need to do this after installation:

source /home/vagrant/.rvm/scripts/rvm
should see the right response after which rvm, which fluentd, which ruby...


Some history:


 185  vi fluent.conf
  186  ls
  187  demosvc -port 3000
  188  demosvc -port 3000 > demosvc1.log&
  189  demosvc -port 3100 > demosvc2.log&
  190  tail demosvc1.log 
  191  ls
  192  unset http_proxy
  193  unset https_proxy
  194  unset HTTP_PROXY
  195  unset HTTPS_PROXY
  196  docker ps -a
  197  docker start 44dcc76d82da
  198  docker start 8deefd9541a9
  199  8909f96483a9
  200  docker start 8909f96483a9
  201  docker ps -a
  202  pwd
  203  ls
  204  fluentd -v -c ./fluent.conf 
  205  pwd
  206  ls
  207  history
  208  which ruby
  209  which fluentd
  210  ls
  211  cd ..
  212  ls
  213  more setup-fluentd.sh 
  214  echo "source /home/vagrant/.rvm/scripts/rvm"
  215  which fluentd
  216  more /home/vagrant/.rvm/scripts/rvm
  217  ls
  218  fluentd
  219  ruby --version
  220  rvm --default use 2.2.0
  221  history
  222  which rvm
  223  source /home/vagrant/.rvm/scripts/rvm
  224  which rvm
  225  which fluentd
  226  which ruby
  227  cd logging
  228  ls
  229  history
  230  ls
  231  history
  232  fluentd -v -c ./fluent.conf
  233  history
  
 vagrant@vagrant-ubuntu-trusty-64:~/logging$ cat fluent.conf 
<source>
  type tail
  path ./demosvc1.log
  pos_file ./demo1-log-pos
  tag demo.json 
  format json 
  time_key fake
</source>
				
<match **>
  type copy
  <store>
  type elasticsearch
  logstash_format true
  host 172.20.20.70
  port 9200
  flush_interval 10s
  </store>
</match>

