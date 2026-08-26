INSERT OR IGNORE INTO users(id,email,password_hash,role,active,created_at) VALUES (1,'mayor@example.com','village-secret','village_admin',1,'2026-01-01T00:00:00Z');
INSERT OR IGNORE INTO users(id,email,password_hash,role,active,created_at) VALUES (2,'coach@example.com','coach-secret','coach',1,'2026-01-01T00:00:00Z');
INSERT OR IGNORE INTO venues(id,name,sport,capacity,active) VALUES (1,'North Arena','mobility',8,1),(2,'Sky Lab','balance',4,1);
INSERT OR IGNORE INTO rooms(id,code,capacity,occupied,version) VALUES (1,'A-101',4,0,1),(2,'A-102',2,0,1),(3,'B-201',6,0,1);
