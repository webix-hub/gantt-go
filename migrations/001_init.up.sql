CREATE TABLE task
(
    id          int auto_increment          primary key,
    text        varchar(2048)               not null,
    start_date  varchar(2048)               not null,
    duration    int                         not null,
    parent      int                         not null,
    progress    float                       not null,
    opened      int                         not null, /* ! */
    details     text    not null
);

CREATE TABLE link
(
    id      int auto_increment  primary key,
    source  int                 not null,
    target  int                 not null,
    type    int default 0       not null
);

INSERT INTO event (id, text, start_date, duration, parent, progress, opened, details) VALUES ("1", "Project A", "2018-06-10 00:00:00", 4, 0, 0.5, 1, "Weekly meeting required\nRoom 508");
INSERT INTO event (id, text, start_date, duration, parent, progress, opened, details) VALUES ("1.1", "Task A1", "2018-06-10 00:00:00", 1, 0.9, "1");
INSERT INTO event (id, text, start_date, duration, parent, progress, opened, details) VALUES ("1.2", "Task A2", "2018-06-11 00:00:00", 3, 0.2, "1", "Contact Rob for details");
INSERT INTO event (id, text, start_date, duration, parent, progress, opened, details) VALUES ("2", "Project B", "2018-06-12 00:00:00", 2, 0, 0);

INSERT INTO link (id, source, target, type) VALUES (1, 1, 2, 1);
