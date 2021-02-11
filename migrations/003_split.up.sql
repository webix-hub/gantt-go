ALTER TABLE task ADD render varchar(5) default '' NOT NULL;
ALTER TABLE task MODIFY COLUMN position INT default 0 NOT NULL;
ALTER TABLE task MODIFY COLUMN type varchar(255) default 'task' NOT NULL;

INSERT INTO task (id, text, start_date, type, duration, parent, progress, opened, details, position, render) VALUES (5, "Project C", "2018-06-10 00:00:00", "project", 2, 0, 0, 0, '', 2, "split");
INSERT INTO task (id, text, start_date, type, duration, parent, progress, opened, details, position) VALUES (6, "Task C1", "2018-06-10 00:00:00", "task", 1, 5, 0, 0, '', 0);
INSERT INTO task (id, text, start_date, type, duration, parent, progress, opened, details, position) VALUES (7, "Task C2", "2018-06-11 00:00:00", "task", 1, 5, 0, 0, '', 1);
INSERT INTO task (id, text, start_date, type, duration, parent, progress, opened, details, position, render) VALUES (8, "Task D", "2018-06-11 00:00:00", "task", 3, 0, 0, 0, '', 3, "split");
INSERT INTO task (id, text, start_date, type, duration, parent, progress, opened, details, position) VALUES (9, "Task D1", "2018-06-11 00:00:00", "task", 1, 8, 0, 0, '', 0);
INSERT INTO task (id, text, start_date, type, duration, parent, progress, opened, details, position) VALUES (10, "Task D2", "2018-06-13 00:00:00", "task", 1, 8, 0, 0, '', 1);

INSERT INTO link (id, source, target, type) VALUES (2, 9, 10, 0);
