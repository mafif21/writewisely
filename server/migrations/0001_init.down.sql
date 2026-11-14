-- 0001_init (down): drop the initial schema in reverse dependency order.

BEGIN;

DROP TABLE IF EXISTS evaluation;
DROP TABLE IF EXISTS attempt;
DROP TABLE IF EXISTS writing_submission;

COMMIT;
