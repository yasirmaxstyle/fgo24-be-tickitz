DELETE FROM payment_method
WHERE
    name IN (
        'Google Pay',
        'Visa',
        'Gopay',
        'Ovo',
        'Paypal',
        'BRI',
        'BCA'
    );