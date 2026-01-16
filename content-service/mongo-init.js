db = db.getSiblingDB(process.env.MONGO_INITDB_DATABASE);
db.createUser({
	user: process.env.SRV_CONTENT_MONGO_USERNAME,
	pwd: process.env.SRV_CONTENT_MONGO_PASSWORD,
	roles: [{ role: "readWrite", db: db.getName() }]
});