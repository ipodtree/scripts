## change color schema
export PS1='\[\033[01;32m\]\u@\h\[\033[00m\]:\[\033[01;36m\]\w\[\033[00m\] \D{%F %T}\n\$ '

## system variable
export JAVA_HOME=$(/usr/libexec/java_home)
export ANT_HOME="/Applications/apache-ant-1.10.1"
export PATH="$PATH:$JAVA_HOME/bin:$ANT_HOME/bin"
export M2_HOME="/Applications/apache-maven-3.5.0"
export PATH="$PATH:$M2_HOME/bin"
export PENTAHO_JAVA_HOME=$(/usr/libexec/java_home)

# alias
alias lt='ls -lthr'
alias vps='ssh root@45.77.7.69'
alias www='ssh www@45.77.7.69'
alias vm='ssh root@vm-qiading-001'
alias lae='ssh qiading@laeusr-prod2-01'
