-- Modify "addresses" table
ALTER TABLE "public"."addresses" DROP CONSTRAINT "fk_users_addresses";
-- Modify "users" table
ALTER TABLE "public"."users" DROP COLUMN "avatar", ADD COLUMN "user_name" text NOT NULL, ADD COLUMN "email" text NOT NULL, ADD COLUMN "phone" text NOT NULL, ADD CONSTRAINT "uni_users_email" UNIQUE ("email"), ADD CONSTRAINT "uni_users_phone" UNIQUE ("phone"), ADD CONSTRAINT "uni_users_user_name" UNIQUE ("user_name");
