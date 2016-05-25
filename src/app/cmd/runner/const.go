package runner

const (
	WorkingDir    = "/opt/postgres-ci/build/"
	AssetsDir     = "/opt/postgres-ci/assets/"
	TestRunnerSql = `
		SELECT 
			namespace,
			procedure,
			to_json(errors) AS errors,
			started_at,
			finished_at
		FROM assert.test_runner()
	`
)
