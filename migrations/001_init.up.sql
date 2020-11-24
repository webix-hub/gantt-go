CREATE TABLE task
(
    id          int auto_increment          primary key,
    text        varchar(2048)               not null,
    start_date  varchar(2048)               not null,
    type        varchar(255)                not null,
    duration    int                         not null,
    parent      int                         not null,
    progress    float                       not null,
    opened      TINYINT(1) default 0        not null, /* ! */
    details     text                        not null  
);

CREATE TABLE link
(
    id      int auto_increment  primary key,
    source  int                 not null,
    target  int                 not null,
    type    int default 0       not null
);

INSERT INTO task (id, text, start_date, type, duration, parent, progress, opened, details) VALUES (1, "Project A", "2018-06-10 00:00:00", "project", 4, 0, 0.5, 1, "Weekly meeting required\nRoom 508");
INSERT INTO task (id, text, start_date, type, duration, parent, progress, opened, details) VALUES (2, "Task A1", "2018-06-10 00:00:00", "task", 1, 1, 0.9, 0, '');
INSERT INTO task (id, text, start_date, type, duration, parent, progress, opened, details) VALUES (3, "Task A2", "2018-06-11 00:00:00", "task", 3, 1, 0.2, 0, "Contact Rob for details");
INSERT INTO task (id, text, start_date, type, duration, parent, progress, opened, details) VALUES (4, "Project B", "2018-06-12 00:00:00", "project", 2, 0, 0, 0, '');

INSERT INTO link (id, source, target, type) VALUES (1, 1, 2, 1);
