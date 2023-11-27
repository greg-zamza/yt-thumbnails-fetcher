CREATE DATABASE app_db;
\c app_db;

CREATE TABLE admins (
    id INT primary key
);

CREATE TABLE ignore_list (
    id VARCHAR(24) PRIMARY KEY
);

CREATE TABLE next_page (
    id INT PRIMARY KEY,
    token VARCHAR(10),
    refresh DATE
);

-- insert default values
INSERT INTO next_page(id, token, refresh) VALUES(1, '', CURRENT_DATE);
