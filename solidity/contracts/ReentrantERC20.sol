pragma solidity ^0.8.0;
import "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import "./Gravity.sol";

pragma experimental ABIEncoderV2;

// Reentrant evil erc20
contract ReentrantERC20 {
    address state_gravityAddress;

    constructor(address _gravityAddress) public {
        state_gravityAddress = _gravityAddress;
    }

    function transfer(address recipient, uint256 amount) public returns (bool) {
        // _currentValidators, _currentPowers, _currentValsetNonce, _v, _r, _s, _args);(
        address[] memory addresses = new address[](0);
        uint256[] memory uint256s = new uint256[](0);
        Signature[] memory _sigs = new Signature[](0);
        bytes memory bytess = new bytes(0);
        uint256 zero = 0;
        LogicCallArgs memory args;

        {
            args = LogicCallArgs(
                uint256s,
                addresses,
                uint256s,
                addresses,
                address(0),
                bytess,
                zero,
                bytes32(0),
                zero
            );
        }
        
        Gravity(state_gravityAddress).submitLogicCall(
            addresses, 
            uint256s, 
            zero, 
            _sigs,
            args
        );
    }
}
