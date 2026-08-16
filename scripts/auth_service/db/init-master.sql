CREATE USER IF NOT EXISTS 'repl_user'@'%' IDENTIFIED WITH caching_sha2_password BY 'repl_password';
GRANT REPLICATION SLAVE ON *.* TO 'repl_user'@'%';
FLUSH PRIVILEGES;
