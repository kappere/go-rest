CREATE TABLE `oauth_client_details` (
  `id` varchar(255) NOT NULL COMMENT 'client id',
  `secret` varchar(255) NOT NULL COMMENT 'client secret',
  `domain` varchar(255) DEFAULT NULL COMMENT 'domain',
  `user_id` varchar(255) DEFAULT NULL COMMENT 'user id',
  PRIMARY KEY (`id`)
) COMMENT='oauth2 client details';

CREATE TABLE `oauth_access_token` (
  `client_id` varchar(255) NOT NULL COMMENT 'client id',
  `user_id` varchar(255) DEFAULT NULL COMMENT 'user id',
  `redirect_uri` varchar(255) DEFAULT NULL COMMENT 'redirect uri',
  `scope` varchar(255) DEFAULT NULL COMMENT 'scope',
  `code` varchar(255) DEFAULT NULL COMMENT 'code',
  `code_create_at` varchar(255) DEFAULT NULL COMMENT 'code create at',
  `code_expires_in` varchar(255) DEFAULT NULL COMMENT 'code expires in',
  `access` varchar(255) DEFAULT NULL COMMENT 'access',
  `access_create_at` varchar(255) DEFAULT NULL COMMENT 'access create at',
  `access_expires_in` varchar(255) DEFAULT NULL COMMENT 'access expires in',
  `refresh` varchar(255) DEFAULT NULL COMMENT 'refresh',
  `refresh_create_at` varchar(255) DEFAULT NULL COMMENT 'refresh create at',
  `refresh_expires_in` varchar(255) DEFAULT NULL COMMENT 'refresh expires in',
  PRIMARY KEY (`client_id`)
) COMMENT='oauth2 access token';