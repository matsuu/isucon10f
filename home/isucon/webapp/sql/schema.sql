DROP TABLE IF EXISTS `contestants`;
CREATE TABLE `contestants` (
  `id` VARCHAR(255) PRIMARY KEY,
  `password` VARCHAR(255) NOT NULL,
  `team_id` BIGINT,
  `name` VARCHAR(255),
  `student` TINYINT(1) DEFAULT FALSE,
  `staff` TINYINT(1) DEFAULT FALSE,
  `created_at` DATETIME(6) NOT NULL
) ENGINE=InnoDB DEFAULT CHARACTER SET=utf8mb4;

DROP TABLE IF EXISTS `teams`;
CREATE TABLE `teams` (
  `id` BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
  `name` VARCHAR(255) NOT NULL,
  `leader_id` VARCHAR(255),
  `email_address` VARCHAR(255) NOT NULL,
  `invite_token` VARCHAR(255) NOT NULL,
  `withdrawn` TINYINT(1) DEFAULT FALSE,
  `created_at` DATETIME(6) NOT NULL,
  UNIQUE KEY (`leader_id`)
) ENGINE=InnoDB DEFAULT CHARACTER SET utf8mb4;

DROP TABLE IF EXISTS `benchmark_jobs`;
CREATE TABLE `benchmark_jobs` (
  `id` BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
  `team_id` BIGINT NOT NULL,
  `status` INT NOT NULL,
  `target_hostname` VARCHAR(255) NOT NULL,
  `handle` VARCHAR(255),
  `score_raw` INT,
  `score_deduction` INT,
  `reason` VARCHAR(255),
  `passed` TINYINT(1),
  `started_at` DATETIME(6),
  `finished_at` DATETIME(6),
  `created_at` DATETIME(6) NOT NULL,
  `updated_at` DATETIME(6) NOT NULL
) ENGINE=InnoDB DEFAULT CHARACTER SET=utf8mb4;

ALTER TABLE `benchmark_jobs` ADD INDEX idx1 (`team_id`,`id`);
ALTER TABLE `benchmark_jobs` ADD INDEX idx2 (`status`,`team_id`,`id`);
ALTER TABLE `benchmark_jobs` ADD INDEX idx3 (`status`,`team_id`,`finished_at`);

DROP TABLE IF EXISTS `clarifications`;
CREATE TABLE `clarifications` (
  `id` BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
  `team_id` BIGINT NOT NULL,
  `disclosed` TINYINT(1),
  `question` VARCHAR(255),
  `answer` VARCHAR(255),
  `answered_at` DATETIME(6),
  `created_at` DATETIME(6) NOT NULL,
  `updated_at` DATETIME(6) NOT NULL
) ENGINE=InnoDB DEFAULT CHARACTER SET=utf8mb4;

DROP TABLE IF EXISTS `notifications`;
CREATE TABLE `notifications` (
  `id` BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
  `contestant_id` VARCHAR(255) NOT NULL,
  `read` TINYINT(1) NOT NULL DEFAULT FALSE,
  `encoded_message` VARCHAR(255) NOT NULL,
  `created_at` DATETIME(6) NOT NULL,
  `updated_at` DATETIME(6) NOT NULL
) ENGINE=InnoDB DEFAULT CHARACTER SET=utf8mb4;

ALTER TABLE `notifications` ADD INDEX idx1 (`contestant_id`,`id`);

DROP TABLE IF EXISTS `push_subscriptions`;
CREATE TABLE `push_subscriptions` (
  `id` BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
  `contestant_id` VARCHAR(255) NOT NULL,
  `endpoint` VARCHAR(255) NOT NULL,
  `p256dh` VARCHAR(255) NOT NULL,
  `auth` VARCHAR(255) NOT NULL,
  `created_at` DATETIME(6) NOT NULL,
  `updated_at` DATETIME(6) NOT NULL,
  UNIQUE KEY (`contestant_id`, `endpoint`)
) ENGINE=InnoDB DEFAULT CHARACTER SET=utf8mb4;

DROP TABLE IF EXISTS `contest_config`;
CREATE TABLE `contest_config` (
  `registration_open_at` DATETIME(6) NOT NULL,
  `contest_starts_at` DATETIME(6) NOT NULL,
  `contest_freezes_at` DATETIME(6) NOT NULL,
  `contest_ends_at` DATETIME(6) NOT NULL
) ENGINE=InnoDB DEFAULT CHARACTER SET=utf8mb4;

ALTER TABLE contestants ADD INDEX (team_id);
ALTER TABLE benchmark_jobs ADD INDEX (team_id, created_at);

DROP TABLE IF EXISTS `scores`;
CREATE TABLE `scores` (
  `team_id` BIGINT NOT NULL,
  `student` BOOLEAN NOT NULL DEFAULT FALSE,
  `best_score` INT NOT NULL DEFAULT 0,
  `best_score_started_at` DATETIME(6),
  `best_score_marked_at` DATETIME(6),
  `latest_score` INT NOT NULL DEFAULT 0,
  `latest_score_started_at` DATETIME(6),
  `latest_score_marked_at` DATETIME(6),
  `finish_count` INT NOT NULL DEFAULT 0,
  `freeze_best_score` INT NOT NULL DEFAULT 0,
  `freeze_best_score_started_at` DATETIME(6),
  `freeze_best_score_marked_at` DATETIME(6),
  `freeze_latest_score` INT NOT NULL DEFAULT 0,
  `freeze_latest_score_started_at` DATETIME(6),
  `freeze_latest_score_marked_at` DATETIME(6),
  `freeze_finish_count` INT NOT NULL DEFAULT 0,
  PRIMARY KEY (team_id)
) ENGINE=InnoDB DEFAULT CHARACTER SET=utf8mb4;


DROP TRIGGER IF EXISTS insert_score;
DROP TRIGGER IF EXISTS update_score;
DROP TRIGGER IF EXISTS update_student;
delimiter //
CREATE TRIGGER insert_score AFTER INSERT ON teams
  FOR EACH ROW
  BEGIN
    INSERT INTO `scores` (`team_id`) VALUES (NEW.id);
  END;//


CREATE TRIGGER update_score AFTER UPDATE ON benchmark_jobs
  FOR EACH ROW
  BEGIN
    IF NEW.finished_at IS NOT NULL THEN
      UPDATE scores SET latest_score=NEW.score_raw-NEW.score_deduction, latest_score_started_at=NEW.started_at, latest_score_marked_at=NEW.finished_at, finish_count=finish_count+1 WHERE team_id=NEW.team_id;
      UPDATE scores SET best_score=latest_score, best_score_started_at=latest_score_started_at, best_score_marked_at=latest_score_marked_at WHERE team_id=NEW.team_id AND best_score<=latest_score;
      UPDATE scores SET freeze_best_score=best_score, freeze_best_score_started_at=best_score_started_at, freeze_best_score_marked_at=best_score_marked_at, freeze_latest_score=latest_score, freeze_latest_score_started_at=latest_score_started_at, freeze_latest_score_marked_at=latest_score_marked_at, freeze_finish_count=freeze_finish_count+1 WHERE team_id=NEW.team_id AND latest_score_marked_at<(SELECT MAX(contest_freezes_at) FROM contest_config);
    END IF;
  END;//

CREATE TRIGGER update_student AFTER UPDATE ON contestants
  FOR EACH ROW
  BEGIN
   UPDATE `scores` SET student = (SELECT SUM(student) = COUNT(*) FROM contestants WHERE team_id = NEW.team_id) WHERE team_id = NEW.team_id;
  END;//
delimiter ;

