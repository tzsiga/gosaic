package database

import "database/sql"

type MigrationFunc func(db *sql.DB) error
type Migrations []MigrationFunc

var (
	migrations Migrations = Migrations{
		createVersionTable,
		createAspectTable,
		createGidxTable,
		createGidxPartialTable,
		createCoverTable,
		createCoverPartialTable,
		createMacroTable,
		createMacroPartialTable,
	}
)

func Migrate(db *sql.DB) (int, error) {
	version, err := GetVersion(db)
	if err != nil {
		return version, err
	}

	for idx, migFun := range migrations {
		migVer := idx + 1
		if version < migVer {
			err = migFun(db)
			if err != nil {
				return version, err
			}
			err = setVersion(db, migVer)
			if err != nil {
				return version, err
			}
		}
	}

	version, err = GetVersion(db)
	if err != nil {
		return version, err
	} else {
		return version, nil
	}
}

func GetVersion(db *sql.DB) (int, error) {
	var version int
	sql := `
    select version
    from versions
    order by version desc
    limit 1
  `
	rows, err := db.Query(sql)
	if err != nil {
		// db has not been created yet
		return 0, nil
	}
	defer rows.Close()

	for rows.Next() {
		rows.Scan(&version)
	}
	err = rows.Close()

	return version, err
}

func setVersion(db *sql.DB, version int) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	sql := "insert into versions(version) values(?)"
	stmt, err := tx.Prepare(sql)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(version)
	if err != nil {
		return err
	}
	tx.Commit()
	return nil
}

func createVersionTable(db *sql.DB) error {
	sql := `
    create table versions (
      version integer not null primary key
    );
  `
	_, err := db.Exec(sql)

	if err != nil {
		return err
	}

	sql = "create unique index idx_versions_version on versions (version)"
	_, err = db.Exec(sql)
	return err
}

func createGidxTable(db *sql.DB) error {
	sql := `
    create table gidx (
      id integer not null primary key,
      aspect_id integer not null,
      path text not null,
      md5sum text not null,
      width integer not null,
      height integer not null,
      orientation integer not null,
			FOREIGN KEY(aspect_id) REFERENCES aspects(id) ON DELETE RESTRICT
    );
  `
	_, err := db.Exec(sql)
	if err != nil {
		return err
	}

	sql = "create unique index idx_gidx_md5sum on gidx (md5sum);"
	_, err = db.Exec(sql)
	return err
}

func createAspectTable(db *sql.DB) error {
	sql := `
    create table aspects (
      id integer not null primary key,
      columns integer not null,
      rows integer not null
    );
  `
	_, err := db.Exec(sql)
	if err != nil {
		return err
	}

	sql = "create unique index idx_aspects on aspects (rows,columns);"
	_, err = db.Exec(sql)
	return err
}

func createGidxPartialTable(db *sql.DB) error {
	sql := `
    create table gidx_partials (
      id integer not null primary key,
      gidx_id integer not null,
      aspect_id integer not null,
			data blob not null,
			FOREIGN KEY(gidx_id) REFERENCES gidx(id) ON DELETE CASCADE,
			FOREIGN KEY(aspect_id) REFERENCES aspects(id) ON DELETE RESTRICT
    );
  `
	_, err := db.Exec(sql)
	if err != nil {
		return err
	}

	sql = "create unique index idx_gidx_partials on gidx_partials (gidx_id,aspect_id);"
	_, err = db.Exec(sql)
	return err
}

func createCoverTable(db *sql.DB) error {
	sql := `
    create table covers (
      id integer not null primary key,
      aspect_id integer not null,
			type text not null,
      name text not null,
      width integer not null,
			height integer not null,
			FOREIGN KEY(aspect_id) REFERENCES aspects(id) ON DELETE RESTRICT
    );
  `
	_, err := db.Exec(sql)
	if err != nil {
		return err
	}

	sql = "create unique index idx_covers_name on covers (name);"
	_, err = db.Exec(sql)
	return err
}

func createCoverPartialTable(db *sql.DB) error {
	sql := `
    create table cover_partials (
      id integer not null primary key,
      cover_id integer not null,
      aspect_id integer not null,
			x1 integer not null,
			y1 integer not null,
			x2 integer not null,
			y2 integer not null,
			FOREIGN KEY(cover_id) REFERENCES covers(id) ON DELETE CASCADE,
			FOREIGN KEY(aspect_id) REFERENCES aspects(id) ON DELETE RESTRICT
    );
  `
	_, err := db.Exec(sql)
	if err != nil {
		return err
	}

	sql = "create index idx_cover_partials on cover_partials (cover_id);"
	_, err = db.Exec(sql)
	return err
}

func createMacroTable(db *sql.DB) error {
	sql := `
    create table macros (
      id integer not null primary key,
      aspect_id integer not null,
			cover_id integer not null,
      path text not null,
      md5sum text not null,
      width integer not null,
      height integer not null,
      orientation integer not null,
			FOREIGN KEY(aspect_id) REFERENCES aspects(id) ON DELETE RESTRICT,
			FOREIGN KEY(cover_id) REFERENCES covers(id) ON DELETE RESTRICT
    );
  `
	_, err := db.Exec(sql)
	if err != nil {
		return err
	}

	sql = "create unique index idx_macro_md5sum on macros (md5sum);"
	_, err = db.Exec(sql)
	return err
}

func createMacroPartialTable(db *sql.DB) error {
	sql := `
    create table macro_partials (
      id integer not null primary key,
			macro_id integer not null,
			cover_partial_id integer not null,
      aspect_id integer not null,
			data blob not null,
			FOREIGN KEY(macro_id) REFERENCES macros(id) ON DELETE CASCADE,
			FOREIGN KEY(cover_partial_id) REFERENCES cover_partials(id) ON DELETE CASCADE,
			FOREIGN KEY(aspect_id) REFERENCES aspects(id) ON DELETE RESTRICT
    );
  `
	_, err := db.Exec(sql)
	if err != nil {
		return err
	}

	sql = "create index idx_macro_partial_macro_id on macro_partials (macro_id);"
	_, err = db.Exec(sql)
	if err != nil {
		return err
	}

	sql = "create unique index idx_macro_partial on macro_partials (macro_id,cover_partial_id);"
	_, err = db.Exec(sql)
	return err
}
