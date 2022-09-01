CREATE TABLE `user` (
    `user_id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
    `user_name` varchar(256) DEFAULT NULL,
    `nick_name` varchar(256) DEFAULT NULL,
    `password_digest` varchar(256) DEFAULT NULL,
    PRIMARY KEY (`user_id`),
    UNIQUE KEY `user_name` (`user_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;