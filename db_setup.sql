CREATE DATABASE courses_db;
USE courses_db;
CREATE TABLE courses(ID varchar(100) NOT NULL PRIMARY KEY, Title varchar(100), Instructor varchar(100), Faculty varchar(100));

CREATE DATABASE login_db;
USE login_db;
CREATE TABLE login(Username varchar(100) NOT NULL PRIMARY KEY, Pw varchar(100) NOT NULL, AccessKey varchar(100));