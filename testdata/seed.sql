CREATE TABLE graphql_records (
  id serial primary key
);

INSERT INTO graphql_records (id) VALUES
  (3),
	(1),
	(2),
	(4),
	(5),
	(6),
	(7),
	(8),
	(19),
	(9),
	(10),
	(11),
	(12),
	(14),
	(15),
	(16),
	(17),
	(18),
	(13),
	(20);

alter sequence graphql_records_id_seq restart with 100;
