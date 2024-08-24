TICKER = "PLTR"

users = [
    {
        "Id": 1,
        "Balance": {
            "PLTR": 10,
            "USD": 50000,
        },
    },
    {
        "Id": 2,
        "Balance": {
            "PLTR": 10,
            "USD": 50000,
        },
    },
]

bids = []
asks = []


def findUser(userId):
    for user in users:
        if user["Id"] == userId:
            return user
    return None


def flipBalance(userId1, userId2, quantity, price):
	user1 = findUser(userId1)
	user2 = findUser(userId2)
    
	user1["Balance"][TICKER] -= int(quantity)
	user2["Balance"][TICKER] += int(quantity)
    
	user1["Balance"]["UDS"] -= price * quantity
	user2["Balance"]["USD"] += price * quantity


def fillOrder(side, price, quantity, userId):
    remainingQuantity = quantity

    if side == "bid":
        for i in range(len(asks) - 1, -1, -1):
            if asks["price"] > price:
                continue
            if asks[i]["quantity"] > remainingQuantity:
                asks[i]["quantity"] -= remainingQuantity
                flipBalance(asks[i]["userId"], userId, remainingQuantity, price)
                return 0
            else:
                remainingQuantity -= asks[i]["quantity"]
                flipBalance(asks[i]["userId"], userId, asks[i]["quantity"], price)
                asks.pop()
    else:
        for i in range(len(bids) - 1, -1, -1):
            if bids[i]["price"] < price:
                continue
            if bids[i]["quantity"] > remainingQuantity:
                bids[i]["quantity"] -= remainingQuantity
                flipBalance(userId, bids[i]["userId"], remainingQuantity, price)
                return 0
            else:
                remainingQuantity -= bids[i]["quantity"]
                flipBalance(userId, bids[i]["userId"], bids[i]["quantity"], price)
                bids.pop()
        return remainingQuantity
                
            
def handleOrder(userId, side, quantity, price):
    remainingQuantity = fillOrder(side, price, quantity, userId)
    if remainingQuantity == 0:
        return
    else:
        order = {"userId": userId, "price": price, "quantity": remainingQuantity}
        if side == "bid":
            bids.append(order)
            bids.sort(key=lambda x: x['price'])
        else:
            asks.append(order)
            asks.sort(key=lambda x: x['price'], reverse=True)
            

def findDepth():
    depth = bids + asks
    return depth


def balanceOf(userId):
    user = findUser(userId)
    PLTRQuantity = user["Balance"]["PLTR"]
    USDQuantity = user["Balance"]["USD"]
    return "User ID: " + str(userId) + " PLTR Quantity: " + str(PLTRQuantity) + " USD: " + str(USDQuantity)


while True:
    operation = int(input("Enter Operation: "))
    
    if operation == 1:
        userId = int(input("Enter user id: "))
        side = (input("Enter Side: "))
        quantity = int(input("Enter Quantity: "))
        price = int(input("Enter Price: "))
        handleOrder(userId, side, quantity, price)
    elif operation == 2:
        print(findDepth())
    elif operation == 3:
        userId = int(input("Enter user id: "))
        print(balanceOf(userId))
    else:
        print("Invalid Operation")
