import { useEffect, useState } from 'react';

const CountdownTimer = ({ targetTime }) => {
    const calculateTimeLeft = () => {
        const difference = new Date(targetTime).getTime() - new Date().getTime();
        if (difference <= 0) return null;

        const hours = Math.floor(difference / (1000 * 60 * 60));
        const minutes = Math.floor((difference / (1000 * 60)) % 60);
        const seconds = Math.floor((difference / 1000) % 60);

        return { hours, minutes, seconds };
    };

    const [timeLeft, setTimeLeft] = useState(calculateTimeLeft());

    useEffect(() => {
        const timer = setInterval(() => {
            setTimeLeft(calculateTimeLeft());
        }, 1000);

        return () => clearInterval(timer);
    }, [targetTime]);

    if (!timeLeft) {
        return <p className="text-red-500">Expired!</p>;
    }

    return (
        <p className="text-gray-800 font-medium">
            {String(timeLeft.hours).padStart(2, '0')}:
            {String(timeLeft.minutes).padStart(2, '0')}:
            {String(timeLeft.seconds).padStart(2, '0')}
        </p>
    );
};

export default CountdownTimer;
