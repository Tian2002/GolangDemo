create database `xml`;

use `xml`;

drop table `go_xml`;

create table `go_xml`(
	`id` int primary key auto_increment,
    `name` varchar(40),
    `sex` enum('M','W'),
    `phone` varchar(18),
    `created_at` datetime,
    `updated_at` datetime,
    `deleted_at` datetime
);

select* from `go_xml`;
