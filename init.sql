CREATE DATABASE app_db;
\c app_db;

CREATE TABLE admins (
    id INT primary key
);

CREATE TABLE ignore_list (
    id VARCHAR(24) PRIMARY KEY
);

CREATE TABLE next_page (
    token VARCHAR(10) PRIMARY KEY,
    refresh DATE
);
