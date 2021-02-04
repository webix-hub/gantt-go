CREATE TABLE assignment
(
    id          int auto_increment      primary KEY,
    task        int                     NOT NULL,
    resource    int                     NOT NULL,
    value       int                     NOT NULL
);

CREATE TABLE resource
(
    id      INT auto_increment              PRIMARY KEY,
    text    VARCHAR(128)                    NOT NULL,
    parent  INT DEFAULT 0                   NOT NULL,
    avatar  VARCHAR(1024) DEFAULT ''        NOT NULL,
    unit    varchar(32) DEFAULT 'hour'      NOT NULL
) CHARSET=utf8mb4;

INSERT INTO assignment (id, task, resource, value) VALUES (1, 3, 3, 4);
INSERT INTO assignment (id, task, resource, value) VALUES (2, 3, 4, 8);

INSERT INTO resource (id, text, parent, avatar, unit) VALUES (1, "QA", 0, '', "hour");
INSERT INTO resource (id, text, parent, avatar, unit) VALUES (2, "Development", 0, '', "hour");
INSERT INTO resource (id, text, parent, avatar, unit) VALUES (3, "John", 1, "https://docs.webix.com/usermanager-backend/users/101/avatar/092352563.jpg", '');
INSERT INTO resource (id, text, parent, avatar, unit) VALUES (4, "Mike", 2, '', '');
INSERT INTO resource (id, text, parent, avatar, unit) VALUES (5, "Anna Meyer", 2, "https://docs.webix.com/usermanager-backend/users/98/avatar/909471384.jpg", '');
INSERT INTO resource (id, text, parent, avatar, unit) VALUES (6, "Alexander", 2, "https://docs.webix.com/usermanager-backend/users/102/avatar/898151818.jpg", '');
INSERT INTO resource (id, text, parent, avatar, unit) VALUES (7, "Mark", 1, '', '');
INSERT INTO resource (id, text, parent, avatar, unit) VALUES (8, "Leonard", 1, '', '');
INSERT INTO resource (id, text, parent, avatar, unit) VALUES (9, "Design", 0, '', "hour");
INSERT INTO resource (id, text, parent, avatar, unit) VALUES (10, "Alina", 9, '', '');
INSERT INTO resource (id, text, parent, avatar, unit) VALUES (11, "Stephan", 9, '', '');
