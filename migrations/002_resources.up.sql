CREATE TABLE assignment
(
    id          int auto_increment      PRIMARY KEY,
    task        int                     NOT NULL,
    resource    int                     NOT NULL,
    value       int                     NOT NULL
);

CREATE TABLE resource
(
    id      INT auto_increment              PRIMARY KEY,
    name    VARCHAR(128)                    NOT NULL,
    department  INT DEFAULT 0               NOT NULL,
    avatar  VARCHAR(1024) DEFAULT ''        NOT NULL,
    unit    VARCHAR(32) DEFAULT 'hour'      NOT NULL
) CHARSET=utf8mb4;

CREATE TABLE department
(
    id      INT auto_increment          PRIMARY KEY,
    name    VARCHAR(128)                NOT NULL,
    unit    VARCHAR(32) default 'hour'  NOT NULL
) CHARSET=utf8mb4;

INSERT INTO assignment (id, task, resource, value) VALUES (1, 3, 3, 4);
INSERT INTO assignment (id, task, resource, value) VALUES (2, 3, 4, 8);

INSERT INTO resource (id, name, department, avatar, unit) VALUES (1, "John", 1, "https://docs.webix.com/usermanager-backend/users/101/avatar/092352563.jpg", '');
INSERT INTO resource (id, name, department, avatar, unit) VALUES (2, "Mike", 2, '', '');
INSERT INTO resource (id, name, department, avatar, unit) VALUES (3, "Anna Meyer", 2, "https://docs.webix.com/usermanager-backend/users/98/avatar/909471384.jpg", '');
INSERT INTO resource (id, name, department, avatar, unit) VALUES (4, "Alexander", 2, "https://docs.webix.com/usermanager-backend/users/102/avatar/898151818.jpg", '');
INSERT INTO resource (id, name, department, avatar, unit) VALUES (5, "Mark", 1, '', '');
INSERT INTO resource (id, name, department, avatar, unit) VALUES (6, "Leonard", 1, '', '');
INSERT INTO resource (id, name, department, avatar, unit) VALUES (7, "Alina", 3, '', '');
INSERT INTO resource (id, name, department, avatar, unit) VALUES (8, "Stephan", 3, '', '');

INSERT INTO department (id, name, unit) VALUES (1, "QA", "hour");
INSERT INTO department (id, name, unit) VALUES (2, "Development", "hour");
INSERT INTO department (id, name, unit) VALUES (3, "Design", "hour");