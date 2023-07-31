-- -----------------------------------------------------
-- Schema bitespeed
-- -----------------------------------------------------
CREATE SCHEMA IF NOT EXISTS `bitespeed` ;
USE `bitespeed` ;

-- -----------------------------------------------------
-- Table `bitespeed`.`contact`
-- -----------------------------------------------------
CREATE TABLE IF NOT EXISTS contact (
  id SERIAL PRIMARY KEY,
  phone_number VARCHAR(100),
  email VARCHAR(100),
  linked_id INT,
  link_precedence VARCHAR(20) NOT NULL CHECK (link_precedence IN ('primary', 'secondary')),
  created_at TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP NULL
);
