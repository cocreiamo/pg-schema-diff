package migration_acceptance_tests

import (
	"testing"

	"github.com/stripe/pg-schema-diff/pkg/diff"
)

// postgres18VersionNum is the first server_version_num that supports WITHOUT OVERLAPS on primary keys and unique
// constraints.
const postgres18VersionNum = 180000

var exclusionConstraintAcceptanceTestCases = []acceptanceTestCase{
	{
		name: "No-op",
		oldSchemaDDL: []string{
			`
            CREATE EXTENSION btree_gist;
            CREATE TABLE foobar(
                id INT NOT NULL,
                validity TSTZRANGE NOT NULL,
                active BOOLEAN NOT NULL,
                CONSTRAINT foobar_no_overlap EXCLUDE USING gist (id WITH =, validity WITH &&),
                CONSTRAINT foobar_active_no_overlap EXCLUDE USING gist (id WITH =, validity WITH &&) WHERE (active)
                    DEFERRABLE INITIALLY DEFERRED
            );
			`,
		},
		newSchemaDDL: []string{
			`
            CREATE EXTENSION btree_gist;
            CREATE TABLE foobar(
                id INT NOT NULL,
                validity TSTZRANGE NOT NULL,
                active BOOLEAN NOT NULL,
                CONSTRAINT foobar_no_overlap EXCLUDE USING gist (id WITH =, validity WITH &&),
                CONSTRAINT foobar_active_no_overlap EXCLUDE USING gist (id WITH =, validity WITH &&) WHERE (active)
                    DEFERRABLE INITIALLY DEFERRED
            );
			`,
		},
		expectEmptyPlan: true,
	},
	{
		name: "Add table with an exclusion constraint",
		oldSchemaDDL: []string{
			`
            CREATE EXTENSION btree_gist;
			`,
		},
		newSchemaDDL: []string{
			`
            CREATE EXTENSION btree_gist;
            CREATE SCHEMA schema_1;
            CREATE TABLE schema_1.foobar(
                id INT NOT NULL,
                validity TSTZRANGE NOT NULL,
                CONSTRAINT foobar_no_overlap EXCLUDE USING gist (id WITH =, validity WITH &&)
            );
			`,
		},
		expectedPlanDDL: []string{
			"CREATE SCHEMA \"schema_1\"",
			"CREATE TABLE \"schema_1\".\"foobar\" (\n\t\"id\" integer NOT NULL,\n\t\"validity\" tstzrange NOT NULL\n)",
			"ALTER TABLE \"schema_1\".\"foobar\" ADD CONSTRAINT \"foobar_no_overlap\" EXCLUDE USING gist (id WITH =, validity WITH &&)",
		},
	},
	{
		name: "Add an exclusion constraint to an existing table",
		oldSchemaDDL: []string{
			`
            CREATE EXTENSION btree_gist;
            CREATE TABLE foobar(
                id INT NOT NULL,
                validity TSTZRANGE NOT NULL
            );
			`,
		},
		newSchemaDDL: []string{
			`
            CREATE EXTENSION btree_gist;
            CREATE TABLE foobar(
                id INT NOT NULL,
                validity TSTZRANGE NOT NULL,
                CONSTRAINT foobar_no_overlap EXCLUDE USING gist (id WITH =, validity WITH &&) WHERE (NOT isempty(validity))
                    DEFERRABLE
            );
			`,
		},
		expectedHazardTypes: []diff.MigrationHazardType{
			diff.MigrationHazardTypeAcquiresAccessExclusiveLock,
		},
		expectedPlanDDL: []string{
			"ALTER TABLE \"public\".\"foobar\" ADD CONSTRAINT \"foobar_no_overlap\" EXCLUDE USING gist (id WITH =, validity WITH &&) WHERE ((NOT isempty(validity))) DEFERRABLE",
		},
	},
	{
		name: "Drop an exclusion constraint",
		oldSchemaDDL: []string{
			`
            CREATE EXTENSION btree_gist;
            CREATE TABLE foobar(
                id INT NOT NULL,
                validity TSTZRANGE NOT NULL,
                CONSTRAINT foobar_no_overlap EXCLUDE USING gist (id WITH =, validity WITH &&)
            );
			`,
		},
		newSchemaDDL: []string{
			`
            CREATE EXTENSION btree_gist;
            CREATE TABLE foobar(
                id INT NOT NULL,
                validity TSTZRANGE NOT NULL
            );
			`,
		},
		expectedHazardTypes: []diff.MigrationHazardType{
			diff.MigrationHazardTypeAcquiresAccessExclusiveLock,
			diff.MigrationHazardTypeIndexDropped,
		},
		expectedPlanDDL: []string{
			"ALTER TABLE \"public\".\"foobar\" DROP CONSTRAINT \"foobar_no_overlap\"",
		},
	},
	{
		name: "Change the operators of an exclusion constraint",
		oldSchemaDDL: []string{
			`
            CREATE EXTENSION btree_gist;
            CREATE TABLE foobar(
                id INT NOT NULL,
                validity TSTZRANGE NOT NULL,
                CONSTRAINT foobar_no_overlap EXCLUDE USING gist (id WITH =, validity WITH &&)
            );
			`,
		},
		newSchemaDDL: []string{
			`
            CREATE EXTENSION btree_gist;
            CREATE TABLE foobar(
                id INT NOT NULL,
                validity TSTZRANGE NOT NULL,
                CONSTRAINT foobar_no_overlap EXCLUDE USING gist (id WITH <>, validity WITH &&)
            );
			`,
		},
		expectedHazardTypes: []diff.MigrationHazardType{
			diff.MigrationHazardTypeAcquiresAccessExclusiveLock,
			diff.MigrationHazardTypeIndexDropped,
		},
	},
	{
		name: "Turn a plain gist index into an exclusion constraint with the same name",
		oldSchemaDDL: []string{
			`
            CREATE EXTENSION btree_gist;
            CREATE TABLE foobar(
                id INT NOT NULL,
                validity TSTZRANGE NOT NULL
            );
            CREATE INDEX foobar_no_overlap ON foobar USING gist (id, validity);
			`,
		},
		newSchemaDDL: []string{
			`
            CREATE EXTENSION btree_gist;
            CREATE TABLE foobar(
                id INT NOT NULL,
                validity TSTZRANGE NOT NULL,
                CONSTRAINT foobar_no_overlap EXCLUDE USING gist (id WITH =, validity WITH &&)
            );
			`,
		},
		expectedHazardTypes: []diff.MigrationHazardType{
			diff.MigrationHazardTypeAcquiresAccessExclusiveLock,
			diff.MigrationHazardTypeIndexDropped,
		},
	},
	{
		name:                      "No-op with WITHOUT OVERLAPS constraints",
		minimumPostgresVersionNum: postgres18VersionNum,
		oldSchemaDDL: []string{
			`
            CREATE EXTENSION btree_gist;
            CREATE TABLE foobar(
                id INT NOT NULL,
                code TEXT NOT NULL,
                validity TSTZRANGE NOT NULL,
                PRIMARY KEY (id, validity WITHOUT OVERLAPS),
                UNIQUE (code, validity WITHOUT OVERLAPS)
            );
			`,
		},
		newSchemaDDL: []string{
			`
            CREATE EXTENSION btree_gist;
            CREATE TABLE foobar(
                id INT NOT NULL,
                code TEXT NOT NULL,
                validity TSTZRANGE NOT NULL,
                PRIMARY KEY (id, validity WITHOUT OVERLAPS),
                UNIQUE (code, validity WITHOUT OVERLAPS)
            );
			`,
		},
		expectEmptyPlan: true,
	},
	{
		name:                      "Add table with a WITHOUT OVERLAPS primary key and unique constraint",
		minimumPostgresVersionNum: postgres18VersionNum,
		oldSchemaDDL: []string{
			`
            CREATE EXTENSION btree_gist;
			`,
		},
		newSchemaDDL: []string{
			`
            CREATE EXTENSION btree_gist;
            CREATE TABLE foobar(
                id INT NOT NULL,
                code TEXT NOT NULL,
                validity TSTZRANGE NOT NULL,
                CONSTRAINT foobar_pkey PRIMARY KEY (id, validity WITHOUT OVERLAPS),
                CONSTRAINT foobar_code_key UNIQUE (code, validity WITHOUT OVERLAPS)
            );
			`,
		},
		expectedPlanDDL: []string{
			"CREATE TABLE \"public\".\"foobar\" (\n\t\"id\" integer NOT NULL,\n\t\"code\" text COLLATE \"pg_catalog\".\"default\" NOT NULL,\n\t\"validity\" tstzrange NOT NULL\n)",
			"ALTER TABLE \"public\".\"foobar\" ADD CONSTRAINT \"foobar_code_key\" UNIQUE (code, validity WITHOUT OVERLAPS)",
			"ALTER TABLE \"public\".\"foobar\" ADD CONSTRAINT \"foobar_pkey\" PRIMARY KEY (id, validity WITHOUT OVERLAPS)",
		},
	},
	{
		name:                      "Add a WITHOUT OVERLAPS unique constraint to an existing table",
		minimumPostgresVersionNum: postgres18VersionNum,
		oldSchemaDDL: []string{
			`
            CREATE EXTENSION btree_gist;
            CREATE TABLE foobar(
                id INT NOT NULL,
                validity TSTZRANGE NOT NULL
            );
			`,
		},
		newSchemaDDL: []string{
			`
            CREATE EXTENSION btree_gist;
            CREATE TABLE foobar(
                id INT NOT NULL,
                validity TSTZRANGE NOT NULL,
                CONSTRAINT foobar_id_validity_key UNIQUE (id, validity WITHOUT OVERLAPS)
            );
			`,
		},
		expectedHazardTypes: []diff.MigrationHazardType{
			diff.MigrationHazardTypeAcquiresAccessExclusiveLock,
		},
		expectedPlanDDL: []string{
			"ALTER TABLE \"public\".\"foobar\" ADD CONSTRAINT \"foobar_id_validity_key\" UNIQUE (id, validity WITHOUT OVERLAPS)",
		},
	},
	{
		name:                      "Drop a WITHOUT OVERLAPS primary key",
		minimumPostgresVersionNum: postgres18VersionNum,
		oldSchemaDDL: []string{
			`
            CREATE EXTENSION btree_gist;
            CREATE TABLE foobar(
                id INT NOT NULL,
                validity TSTZRANGE NOT NULL,
                PRIMARY KEY (id, validity WITHOUT OVERLAPS)
            );
			`,
		},
		newSchemaDDL: []string{
			`
            CREATE EXTENSION btree_gist;
            CREATE TABLE foobar(
                id INT NOT NULL,
                validity TSTZRANGE NOT NULL
            );
			`,
		},
		expectedHazardTypes: []diff.MigrationHazardType{
			diff.MigrationHazardTypeAcquiresAccessExclusiveLock,
			diff.MigrationHazardTypeIndexDropped,
		},
		expectedPlanDDL: []string{
			"ALTER TABLE \"public\".\"foobar\" DROP CONSTRAINT \"foobar_pkey\"",
		},
	},
	{
		name:                      "Replace a primary key and an exclusion constraint with a WITHOUT OVERLAPS primary key",
		minimumPostgresVersionNum: postgres18VersionNum,
		oldSchemaDDL: []string{
			`
            CREATE EXTENSION btree_gist;
            CREATE TABLE foobar(
                id INT NOT NULL,
                validity TSTZRANGE NOT NULL,
                CONSTRAINT foobar_pkey PRIMARY KEY (id),
                CONSTRAINT foobar_no_overlap EXCLUDE USING gist (id WITH =, validity WITH &&)
            );
			`,
		},
		newSchemaDDL: []string{
			`
            CREATE EXTENSION btree_gist;
            CREATE TABLE foobar(
                id INT NOT NULL,
                validity TSTZRANGE NOT NULL,
                CONSTRAINT foobar_pkey PRIMARY KEY (id, validity WITHOUT OVERLAPS)
            );
			`,
		},
		expectedHazardTypes: []diff.MigrationHazardType{
			diff.MigrationHazardTypeAcquiresAccessExclusiveLock,
			diff.MigrationHazardTypeIndexDropped,
		},
	},
	{
		name:                      "Add a temporal foreign key referencing a WITHOUT OVERLAPS primary key",
		minimumPostgresVersionNum: postgres18VersionNum,
		oldSchemaDDL: []string{
			`
            CREATE EXTENSION btree_gist;
			`,
		},
		newSchemaDDL: []string{
			`
            CREATE EXTENSION btree_gist;
            CREATE TABLE parent(
                id INT NOT NULL,
                validity TSTZRANGE NOT NULL,
                PRIMARY KEY (id, validity WITHOUT OVERLAPS)
            );
            CREATE TABLE child(
                id INT NOT NULL,
                parent_id INT NOT NULL,
                validity TSTZRANGE NOT NULL,
                PRIMARY KEY (id, validity WITHOUT OVERLAPS),
                FOREIGN KEY (parent_id, PERIOD validity) REFERENCES parent (id, PERIOD validity)
            );
			`,
		},
	},
}

func TestExclusionConstraintTestCases(t *testing.T) {
	runTestCases(t, exclusionConstraintAcceptanceTestCases)
}
