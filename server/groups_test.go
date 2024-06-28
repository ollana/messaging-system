package main

//func TestCreateGroup(t *testing.T) {
//	ctx := context.Background()
//	dbClient = newMockDBClient()
//	t.Run("Create group successfully", func(t *testing.T) {
//		// create a new request
//		req := createGroupRequest{
//			GroupName: "test-group",
//		}
//
//		// seed the request to the handler
//		resp, err := createGroup(ctx, &req)
//		assert.NoError(t, err)
//		assert.NotEmpty(t, resp.GroupId)
//		assert.Equal(t, req.GroupName, resp.GroupName)
//
//	})
//
//	t.Run("db error", func(t *testing.T) {
//		dbClient = newMockDBClient()
//		dbClient.(*mockDBClient).error = fmt.Errorf("some error")
//		req := createGroupRequest{
//			GroupName: "test-group",
//		}
//		resp, err := createGroup(ctx, &req)
//		assert.Error(t, err)
//		assert.IsType(t, &InternalServerError{}, err)
//		assert.Nil(t, resp)
//
//	})
//
//}
//
//func TestAddUserToGroup(t *testing.T) {
//	ctx := context.Background()
//	dbClient = newMockDBClient()
//
//	t.Run("Add user to group successfully", func(t *testing.T) {
//		dbClient.StoreUser(context.Background(), dbUser{UserId: "test-user-1"})
//		dbClient.StoreGroup(context.Background(), dbGroup{GroupId: "test-group-1"})
//		// create a new request
//		req := userToGroupRequest{
//			UserId: "test-user-1",
//		}
//
//		err := addUserToGroup(ctx, "test-group-1", &req)
//		assert.NoError(t, err)
//
//		user, _ := dbClient.GetUser(ctx, "test-user-1")
//		assert.Contains(t, user.Groups, "test-group-1")
//		group, _ := dbClient.GetGroup(ctx, "test-group-1")
//		assert.Contains(t, group.Members, "test-user-1")
//
//	})
//
//	t.Run("invalid group", func(t *testing.T) {
//		// create a new request
//		dbClient.StoreUser(ctx, dbUser{UserId: "test-user-2"})
//		req := userToGroupRequest{
//			UserId: "test-user-2",
//		}
//		err := addUserToGroup(ctx, "test-group-2", &req)
//		assert.Error(t, err)
//		assert.IsType(t, &NotFoundError{}, err)
//	})
//
//	t.Run("invalid user", func(t *testing.T) {
//		dbClient.StoreGroup(ctx, dbGroup{GroupId: "test-group-3"})
//		// create a new request
//		req := userToGroupRequest{
//			UserId: "test-user-3",
//		}
//		err := addUserToGroup(ctx, "test-group-3", &req)
//		assert.Error(t, err)
//		assert.IsType(t, &NotFoundError{}, err)
//	})
//
//	t.Run("user already in group", func(t *testing.T) {
//		dbClient.StoreUser(ctx, dbUser{UserId: "test-user-4"})
//		dbClient.StoreGroup(ctx, dbGroup{GroupId: "test-group-4"})
//		req := userToGroupRequest{
//			UserId: "test-user-4",
//		}
//		addUserToGroup(ctx, "test-group-4", &req)
//
//		// add the user again
//		err := addUserToGroup(ctx, "test-group-4", &req)
//		assert.Error(t, err)
//		assert.IsType(t, &BadRequestError{}, err)
//	})
//}
//
//func TestRemoveUserFromGroup(t *testing.T) {
//	ctx := context.Background()
//	dbClient = newMockDBClient()
//
//	t.Run("Remove user from group successfully", func(t *testing.T) {
//		dbClient.StoreUser(ctx, dbUser{UserId: "test-user-1"})
//		dbClient.StoreGroup(ctx, dbGroup{GroupId: "test-group-1"})
//		dbClient.AddUserToGroup(ctx, dbGroup{GroupId: "test-group-1"}, dbUser{UserId: "test-user-1"})
//
//		// create a new request
//		req := userToGroupRequest{
//			UserId: "test-user-1",
//		}
//		err := removeUserFromGroup(ctx, "test-group-1", &req)
//		assert.NoError(t, err)
//
//	})
//}
