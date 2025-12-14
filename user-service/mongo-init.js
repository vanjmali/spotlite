db = db.getSiblingDB(process.env.MONGO_INITDB_DATABASE);
db.createUser({
	user: process.env.SRV_USER_MONGO_USERNAME,
	pwd: process.env.SRV_USER_MONGO_PASSWORD,
	roles: [{ role: "readWrite", db: db.getName() }]
});